package backup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-logr/logr"
	api "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/cmd/mybackup/pkg"
	"go.uber.org/zap/zapcore"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type BackupManager struct {
	log           logr.Logger
	client        client.Client
	cluster       *api.MysqlCluster
	mysqlPassword string
	workDir       string
	backupDir     string
	threads       int
	mysqlVersion  string
	// status fields
	startTime   time.Time
	sourceIndex string
	status      pkg.ServerStatus
	gtidSet     string
	warnings    []string
}

func NewBackupManager(cfg *rest.Config, dir, ns, name string, threads int) (*BackupManager, error) {
	ctx := context.Background()
	log := zap.New(zap.WriteTo(os.Stderr), zap.StacktraceLevel(zapcore.DPanicLevel))
	scheme := runtime.NewScheme()
	if err := api.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add scheme: %w", err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add corev1 scheme: %w", err)
	}
	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create controller-runtime client: %w", err)
	}
	cluster := &api.MysqlCluster{}
	if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: ns, Name: name}, cluster); err != nil {
		return nil, fmt.Errorf("failed to get the MySQLCluster: %w", err)
	}

	Secret := &corev1.Secret{}

	secretName := cluster.Name + "-secret"
	if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: ns, Name: secretName}, Secret); err != nil {
		return nil, fmt.Errorf("failed to get the Secret: %w", err)
	}
	mysqlVersion := cluster.Spec.MysqlVersion
	backupPassword := string(Secret.Data["internal-root-password"])
	fullBackupDir := filepath.Join(dir, ns, name)
	return &BackupManager{
		log:           log,
		client:        k8sClient,
		cluster:       cluster,
		mysqlVersion:  mysqlVersion,
		mysqlPassword: backupPassword,
		workDir:       dir,
		backupDir:     fullBackupDir,
		threads:       threads,
	}, nil
}

func (bm *BackupManager) Backup(ctx context.Context) error {
	// choose a follower pod to backup
	cluster := bm.cluster
	pods := &corev1.PodList{}
	if err := bm.client.List(ctx, pods, client.InNamespace(cluster.Namespace), client.MatchingLabels{

		"app.kubernetes.io/instance": cluster.Name,
		// "app.kubernetes.io/name":       "mysql",
		// "app.kubernetes.io/managed-by": "radondb-mysql-operator",
		// "healthy":                      "yes",
	}); err != nil {
		return fmt.Errorf("failed to get follower pod list: %w", err)
	}
	if len(pods.Items) == 0 {
		return fmt.Errorf("no healthy follower pod found")
	}
	// choose a follower pod to backup
	SourceIndex := chooseBackupPod(cluster, pods)

	op, err := pkg.NewOperator(SourceIndex.Status.PodIP, pkg.MySQLPort, pkg.BackupUser, bm.mysqlPassword, bm.threads)
	if err != nil {
		return fmt.Errorf("failed to create db operator: %w", err)

	}
	defer op.Close()

	if err := op.GetServerStatus(ctx, &bm.status); err != nil {
		return fmt.Errorf("failed to get server status: %w", err)
	}
	bm.startTime = time.Now().Local()
	bm.backupDir = filepath.Join(bm.backupDir, SourceIndex.Name, bm.startTime.Format("20060102-150405"))
	bm.log.Info("chosen source",
		"index", SourceIndex,
		"time", bm.startTime.Format("20060102-150405"),
		"uuid", bm.status.UUID,
		"binlog", bm.status.CurrentBinlog)
	if err := bm.backupFull(ctx, op); err != nil {
		return fmt.Errorf("failed to take a full dump: %w", err)
	}

	// dump and upload binlog for the second or later backups
	lastBackup := &bm.cluster.Status.Backup
	if !lastBackup.Time.IsZero() {
		if err := bm.backupBinlog(ctx, op); err != nil {
			bm.log.Error(err, "failed to backup binary logs")
			bm.warnings = append(bm.warnings, fmt.Sprintf("failed to backup binary logs: %v", err))
		}
	}

	elapsed := time.Since(bm.startTime)
	backupDirUseage, err := pkg.DirUsage(bm.workDir)

	if err != nil {
		bm.warnings = append(bm.warnings, fmt.Sprintf("failed to get backup dir usage: %v", err))
		bm.log.Error(err, "failed to get current backup dir usage")
	}
	lastBackupSize, err := pkg.DirUsage(bm.backupDir)
	if err != nil {
		bm.warnings = append(bm.warnings, fmt.Sprintf("failed to get backup dir usage: %v", err))
		bm.log.Error(err, "failed to get backup dir usage")
	}
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		cluster := &api.MysqlCluster{}
		if err := bm.client.Get(ctx, client.ObjectKeyFromObject(bm.cluster), cluster); err != nil {
			return err
		}

		sb := &cluster.Status.Backup
		sb.Time = metav1.NewTime(bm.startTime)
		sb.Elapsed = metav1.Duration{Duration: elapsed}
		sb.SourcePod = SourceIndex.Name
		sb.SourceUUID = bm.status.UUID
		sb.BinlogFilename = bm.status.CurrentBinlog
		sb.GTIDSet = bm.gtidSet
		sb.LastBackupSize = lastBackupSize
		sb.BackupDirUsedSize = backupDirUseage
		sb.Warnings = bm.warnings
		return bm.client.Status().Update(ctx, cluster)
	})
	if err != nil {
		return fmt.Errorf("failed to update MySQLCluster status: %w", err)
	}

	bm.log.Info("backup finished successfully")

	return nil

}

func chooseBackupPod(cluster *api.MysqlCluster, pods *corev1.PodList) *corev1.Pod {
	// choose a follower pod to backup
	pod := corev1.Pod{}
	for _, p := range pods.Items {
		if p.Labels["role"] == "FOLLOWER" {
			pod = p
			break
		} else if p.Labels["role"] == "LEADER" {
			pod = p
			break
		}
	}
	return &pod
}

func (bm *BackupManager) backupFull(ctx context.Context, op pkg.Operator) error {
	dumpDir := bm.backupDir
	if err := os.MkdirAll(dumpDir, 0755); err != nil {
		return fmt.Errorf("failed to make dump directory: %w", err)
	}

	if err := op.FullBackup(ctx, dumpDir); err != nil {
		return fmt.Errorf("failed to take a full dump: %w", err)
	}

	gtid, err := pkg.GetGTIDExecuted(dumpDir)
	if err != nil {
		return fmt.Errorf("failed to get GTID set from the dump: %w", err)
	}
	bm.gtidSet = gtid

	usage, err := pkg.DirUsage(dumpDir)
	if err != nil {
		return fmt.Errorf("failed to calculate dir usage: %w", err)
	}
	bm.log.Info("full dump ", "bytes", usage)

	return nil
}

func (bm *BackupManager) backupBinlog(ctx context.Context, op pkg.Operator) error {
	binlogDir := filepath.Join(bm.backupDir, "binlog")
	if err := os.MkdirAll(binlogDir, 0755); err != nil {
		return fmt.Errorf("failed to make binlog dump directory: %w", err)
	}

	lastBackup := &bm.cluster.Status.Backup
	binlogName := lastBackup.BinlogFilename
	if bm.sourceIndex != lastBackup.SourcePod {
		binlogs, err := op.GetBinlogs(ctx, bm.mysqlVersion)
		if err != nil {
			return fmt.Errorf("failed to list binlog files: %w", err)
		}
		if len(binlogs) == 0 {
			return fmt.Errorf("no binlog files found")
		}
		pkg.SortBinlogs(binlogs)
		binlogName = binlogs[0]
	}

	if err := op.BinlogBackup(ctx, binlogDir, binlogName, lastBackup.GTIDSet); err != nil {
		return fmt.Errorf("failed to take a full dump: %w", err)
	}

	usage, err := pkg.DirUsage(binlogDir)
	if err != nil {
		return fmt.Errorf("failed to calculate dir usage: %w", err)
	}
	bm.log.Info("binlog backup usage (binlog)", "bytes", usage)
	return nil
}

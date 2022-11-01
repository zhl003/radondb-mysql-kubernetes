package restore

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/go-logr/logr"
	api "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/cmd/mybackup/pkg"
	"go.uber.org/zap/zapcore"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type RestoreManager struct {
	log          logr.Logger
	client       client.Client
	scheme       *runtime.Scheme
	namespace    string
	name         string
	password     string
	threads      int
	restorePoint time.Time
	workDir      string
	cluster      *api.MysqlCluster
}

var ErrBadConnection = errors.New("the connection hasn't reflected the latest user's privileges")

func NewRestoreManager(cfg *rest.Config, srcNS, srcName, dir, ns, name string, threads int, restorePoint time.Time) (*RestoreManager, error) {
	ctx := context.Background()
	log := zap.New(zap.WriteTo(os.Stderr), zap.StacktraceLevel(zapcore.DPanicLevel))
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add corev1 scheme: %w", err)
	}
	if err := api.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add api scheme: %w", err)
	}
	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create controller-runtime client: %w", err)
	}

	cluster := &api.MysqlCluster{}
	if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: ns, Name: name}, cluster); err != nil {
		return nil, fmt.Errorf("failed to get the MySQLCluster: %w", err)
	}

	secretName := cluster.Name + "-secret"
	secret := &corev1.Secret{}
	if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: ns, Name: secretName}, secret); err != nil {
		return nil, fmt.Errorf("failed to get the secret: %w", err)
	}
	backupPassword := string(secret.Data["internal-root-password"])

	fullBackupDir := filepath.Join(dir, srcNS, srcName)

	// prefix := calcPrefix(srcNS, srcName)
	return &RestoreManager{
		log:          log,
		client:       k8sClient,
		scheme:       scheme,
		namespace:    ns,
		name:         name,
		cluster:      cluster,
		password:     backupPassword,
		threads:      threads,
		restorePoint: restorePoint,
		workDir:      fullBackupDir,
	}, nil
}

func (rm *RestoreManager) Restore(ctx context.Context) error {
	cluster := rm.cluster
	podName := cluster.Status.Nodes[0].Name

	rm.log.Info("waiting for a pod to become ready", "name", podName)
	var pod *corev1.Pod
	for i := 0; i < 600; i++ {
		select {
		case <-time.After(1 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}

		pod = &corev1.Pod{}
		if err := rm.client.Get(ctx, client.ObjectKey{Namespace: rm.namespace, Name: podName}, pod); err != nil {
			continue
		}

		if pod.Status.PodIP != "" {
			break
		}

	}

	op, err := pkg.NewOperator(pod.Status.PodIP, pkg.MySQLPort, pkg.BackupUser, rm.password, rm.threads)
	if err != nil {
		return fmt.Errorf("failed to create an operator: %w", err)
	}
	defer op.Close()

	// ping the database until it becomes ready
	rm.log.Info("waiting for the mysqld to become ready", "name", podName)
	for i := 0; i < 600; i++ {
		select {
		case <-time.After(1 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}

		if err := op.Ping(); err != nil {
			continue
		}
		st := &pkg.ServerStatus{}
		if err := op.GetServerStatus(ctx, st); err != nil {
			rm.log.Error(err, "failed to get server status")
			// SHOW MASTER STATUS fails due to the insufficient privileges,
			// if this restore process connects a target database before moco-agent grants privileges to moco-admin.
			// In this case, the restore process panics and retries from the beginning.
			panic(ErrBadConnection)
		}
		if !st.SuperReadOnly {
			continue
		}
		break
	}

	// get the latest backup
	dumpKey, binlogKey, backupTime := rm.FindDumpLocation()
	if dumpKey.FsName == "" {
		return fmt.Errorf("no available backup")
	}

	rm.log.Info("restoring from a backup", "dump", dumpKey.FsName, "binlog", binlogKey.FsName)

	if err := op.PrepareRestore(ctx); err != nil {
		return fmt.Errorf("failed to prepare instance for restoration: %w", err)
	}

	if err := op.LoadDump(ctx, dumpKey); err != nil {
		return fmt.Errorf("failed to load dump: %w", err)
	}

	rm.log.Info("loaded dump successfully")

	if !backupTime.Equal(rm.restorePoint) && binlogKey.FsName != "" {
		if err := op.LoadBinlog(ctx, binlogKey, rm.restorePoint); err != nil {
			return fmt.Errorf("failed to apply transactions: %w", err)
		}
		rm.log.Info("applied binlog successfully")
	}

	if err := op.FinishRestore(ctx); err != nil {
		return fmt.Errorf("failed to finalize the restoration: %w", err)
	}

	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		cluster = &api.MysqlCluster{}
		if err := rm.client.Get(ctx, client.ObjectKey{Namespace: rm.namespace, Name: rm.name}, cluster); err != nil {
			return err
		}

		return rm.client.Status().Update(ctx, cluster)
	})
	if err != nil {
		return fmt.Errorf("failed to update MySQLCluster status: %w", err)
	}

	return nil
}

func (rm *RestoreManager) FindDumpLocation() (pkg.FsInfo, pkg.FsInfo, time.Time) {

	var dumpDirList = []pkg.FsInfo{}
	fn := func(path string, d fs.DirEntry, err error) error {
		fs := pkg.FsInfo{
			FsName: d.Name(),
			FsPath: path,
		}
		reg := regexp.MustCompile(`[0-9]{8}-[0-9]{6}$`)
		if d.IsDir() && reg.MatchString(fs.FsName) {
			dumpDirList = append(dumpDirList, fs)
		}

		return nil
	}
	if err := filepath.WalkDir(rm.workDir, fn); err != nil {
		rm.log.Error(err, "failed to walk the backup directory")
	}
	if len(dumpDirList) == 0 {
		rm.log.Error(fmt.Errorf("no dump directory found"), "failed to find dump directory")
	}
	var nearest time.Time
	var nearestDump, nearestBinlog pkg.FsInfo

	for _, key := range dumpDirList {
		bkt, _ := time.Parse(pkg.BackupTimeFormat, key.FsName)
		if bkt.After(rm.restorePoint) {
			nearestBinlog = key
			break
		}
		nearestDump = key
		nearest = bkt
	}

	return nearestDump, nearestBinlog, nearest
}

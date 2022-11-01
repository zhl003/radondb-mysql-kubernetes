package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/radondb/radondb-mysql-kubernetes/cmd/mybackup/pkg"
	"github.com/radondb/radondb-mysql-kubernetes/cmd/mybackup/restore"
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
)

var restoreCmd = &cobra.Command{
	Use:   "restore  source_namespace source_name  dest_namespace dest_name restore_point",
	Short: "restore MySQL data from a backup",
	Long: `Restore MySQL data from a backup.

source_namespace: The source MySQLCluster's namespace.
source_name:      The source MySQLCluster's name.
dest_namespace:        The target MySQLCluster's namespace.
dest_name:             The target MySQLCluster's name.
restore_point:  The point-in-time to restore data.  e.g. 20210523-150423`,
	Args: cobra.ExactArgs(5),
	RunE: func(cmd *cobra.Command, args []string) error {
		maxRetry := 4
		for i := 0; i < maxRetry; i++ {
			if err := runRestore(cmd, args); err != restore.ErrBadConnection {
				return err
			}

			fmt.Fprintf(os.Stderr, "bad connection: retrying...\n")
			time.Sleep(1 * time.Second)
		}

		return nil
	},
}

func runRestore(cmd *cobra.Command, args []string) (e error) {
	defer func() {
		if r := recover(); r != nil {
			if r == restore.ErrBadConnection {
				e = r.(error)
			} else {
				panic(r)
			}
		}
	}()

	srcNamespace := args[0]
	srcName := args[1]
	namespace := args[2]
	name := args[3]

	restorePoint, err := time.Parse(pkg.BackupTimeFormat, args[4])
	if err != nil {
		return fmt.Errorf("invalid restore point %s: %w", args[4], err)
	}

	if err != nil {
		return fmt.Errorf("failed to create a bucket interface: %w", err)
	}

	cfg, err := ctrl.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get config for Kubernetes: %w", err)
	}

	rm, err := restore.NewRestoreManager(cfg,
		srcNamespace, srcName,
		commonArgs.dumpDir,
		namespace, name,
		commonArgs.threads,
		restorePoint)
	if err != nil {
		return fmt.Errorf("failed to create a restore manager: %w", err)
	}
	return rm.Restore(cmd.Context())
}

func init() {
	rootCmd.AddCommand(restoreCmd)
}

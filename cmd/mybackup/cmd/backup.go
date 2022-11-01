package cmd

import (
	"fmt"

	"github.com/radondb/radondb-mysql-kubernetes/cmd/mybackup/backup"
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
)

var backupCmd = &cobra.Command{
	Use:   "backup  namespace clustername",
	Short: "backup a MySQLCluster's data to the specified directory",
	Long: `Backup a MySQLCluster's data.

	namespace: The namespace of the RadonDB MySQL cluster.
	clustername:      The name of the MySQLCluster.`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		namespace := args[0]
		name := args[1]
		cfg, err := ctrl.GetConfig()
		if err != nil {
			return fmt.Errorf("failed to get config for Kubernetes: %w", err)
		}
		bm, err := backup.NewBackupManager(cfg, commonArgs.dumpDir, namespace, name, commonArgs.threads)
		if err != nil {
			return fmt.Errorf("failed to create a backup manager: %w", err)
		}
		return bm.Backup(cmd.Context())

	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
}

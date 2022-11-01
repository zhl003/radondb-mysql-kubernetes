package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var commonArgs struct {
	dumpDir string
	threads int
}

var rootCmd = &cobra.Command{
	Use:     "mybackup",
	Version: "v3.0.0",
	Short:   "backup and restore MySQL data",
	Long:    "Backup and restore MySQL data.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		// mysqlsh command creates some files in $HOME.
		os.Setenv("HOME", commonArgs.dumpDir)
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
func init() {
	pf := rootCmd.PersistentFlags()
	pf.StringVar(&commonArgs.dumpDir, "dump-dir", "/backup", "The writable working directory")
	pf.IntVar(&commonArgs.threads, "threads", 4, "The number of threads to be used")
}

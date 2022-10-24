package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"
)

var commonArgs struct {
	workDir      string
	threads      int
	region       string
	endpointURL  string
	usePathStyle bool
}

var mysqlPassword = os.Getenv("MYSQL_PASSWORD")
var rootCmd = &cobra.Command{
	Use:     "mybackup",
	Version: "v2.1.1",
	Short:   "backup and restore MySQL data",
	Long:    "Backup and restore MySQL data.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		if len(mysqlPassword) == 0 {
			return errors.New("no MYSQL_PASSWORD environment variable")
		}
		if len(commonArgs.endpointURL) > 0 {
			_, err := url.Parse(commonArgs.endpointURL)
			if err != nil {
				return fmt.Errorf("invalid endpoint URL %s: %w", commonArgs.endpointURL, err)
			}
		}

		// mysqlsh command creates some files in $HOME.
		os.Setenv("HOME", commonArgs.workDir)
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
	pf.StringVar(&commonArgs.workDir, "work-dir", "/work", "The writable working directory")
	pf.IntVar(&commonArgs.threads, "threads", 4, "The number of threads to be used")
	pf.StringVar(&commonArgs.region, "region", "", "AWS region")
	pf.StringVar(&commonArgs.endpointURL, "endpoint", "", "S3 API endpoint URL")
	pf.BoolVar(&commonArgs.usePathStyle, "use-path-style", false, "Use path-style S3 API")
}

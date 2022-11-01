package pkg

import (
	"fmt"
	"os"
)

var RadonUsers = []string{
	"radondb_metrics",
	"radondb_operator",
	"radondb_repl",
	"root",
}

func MakeBackupDir(backupdir string) error {
	// Create the backup directory if it does not exist.
	if err := os.MkdirAll(backupdir, 0755); err != nil {
		return fmt.Errorf("failed to create the backup directory: %w", err)
	}
	return nil
}

type FsInfo struct {
	FsName string
	FsPath string
}

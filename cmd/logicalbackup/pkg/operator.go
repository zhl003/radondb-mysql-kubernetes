package pkg

import (
	"context"

	_ "github.com/go-sql-driver/mysql"
	
)

type DbOperator interface {
	FullBackup(ctx context.Context, dir string) error
	BinlogBackup(ctx context.Context, dir string, filterGTID string) error
}

type operator struct {
	db *sqlx.Db
}

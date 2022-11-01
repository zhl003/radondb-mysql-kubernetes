package pkg

import (
	"context"
	sql "database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
)

type Operator interface {
	// Close ust be called when the operator is no longer in use.
	Close()
	// GetServerStatus fills ServerStatus struct.
	GetServerStatus(context.Context, *ServerStatus) error
	FullBackup(ctx context.Context, dir string) error
	BinlogBackup(ctx context.Context, dir string, binlogName, filterGTID string) error
	// GetBinlogs returns a list of binary log files on the mysql instance.
	GetBinlogs(ctx context.Context, mysqlVersion string) ([]string, error)
	Ping() error
	PrepareRestore(ctx context.Context) error
	LoadDump(ctx context.Context, dir FsInfo) error
	LoadBinlog(ctx context.Context, binlogDir FsInfo, restorePoint time.Time) error
	FinishRestore(ctx context.Context) error
}

type operator struct {
	db       *sql.DB
	host     string
	port     int
	user     string
	password string
	threads  int
}

var _ Operator = operator{}

// NewOperator creates an Operator.
func NewOperator(host string, port int, user, password string, threads int) (Operator, error) {
	cfg := mysql.NewConfig()
	cfg.User = user
	cfg.Passwd = password
	cfg.Net = "tcp"
	cfg.Addr = fmt.Sprintf("%s:%d", host, port)
	cfg.InterpolateParams = true
	cfg.ParseTime = true
	cfg.Timeout = 5 * time.Second
	cfg.ReadTimeout = 1 * time.Minute
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", host, err)
	}
	db.SetMaxIdleConns(1)
	db.SetConnMaxIdleTime(30 * time.Second)
	return operator{db, host, port, user, password, threads}, nil
}

func (m operator) Close() {
	m.db.Close()
}

func (m operator) Ping() error {
	return m.db.Ping()
}

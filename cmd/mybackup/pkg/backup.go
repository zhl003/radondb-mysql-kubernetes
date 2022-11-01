package pkg

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
)

func (o operator) FullBackup(ctx context.Context, dir string) error {
	args := []string{
		fmt.Sprintf("mysql://%s@%s:%d", o.user, o.host, o.port),
		"--passwords-from-stdin",
		"--save-passwords=never",
		"-C", "False",
		"--",
		"util",
		"dump-instance",
		dir,
		"--excludeUsers=" + strings.Join(RadonUsers, ","),
		"--threads=" + fmt.Sprint(o.threads),
	}

	cmd := exec.CommandContext(ctx, "mysqlsh", args...)
	cmd.Stdin = strings.NewReader(o.password)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (o operator) GetBinlogs(ctx context.Context, mysqlversion string) ([]string, error) {
	var binlogs57 []showBinaryLogs57
	var binlogs80 []showBinaryLogs80
	rows, err := o.db.QueryContext(ctx, `SHOW BINARY LOGS`)
	if err != nil {
		return nil, fmt.Errorf("failed to show binary logs: %w", err)
	}
	if mysqlversion == "5.7" {
		for rows.Next() {
			var binlog showBinaryLogs57
			if err := rows.Scan(&binlog.LogName, &binlog.FileSize); err != nil {
				return nil, fmt.Errorf("failed to scan binary logs: %w", err)
			}
			binlogs57 = append(binlogs57, binlog)
		}
	} else {
		for rows.Next() {
			var binlog showBinaryLogs80
			if err := rows.Scan(&binlog.LogName, &binlog.FileSize, &binlog.Encrypted); err != nil {
				return nil, fmt.Errorf("failed to scan binary logs: %w", err)
			}
			binlogs80 = append(binlogs80, binlog)
		}
	}

	var r []string
	if mysqlversion == "5.7" {
		for _, binlog := range binlogs57 {
			r = append(r, binlog.LogName)
		}
	} else {
		for _, binlog := range binlogs80 {
			r = append(r, binlog.LogName)
		}
	}
	return r, nil
}

func (o operator) BinlogBackup(ctx context.Context, dir, binlogName, filterGTID string) error {
	args := []string{
		"-h", o.host,
		"--port", fmt.Sprint(o.port),
		"--protocol=tcp",
		"-u", o.user,
		"-p" + o.password,
		"--get-server-public-key",
		"--read-from-remote-master=BINLOG-DUMP-GTIDS",
		"--exclude-gtids=" + filterGTID,
		"-t",
		"--raw",
		"--result-file=" + dir + "/",
		binlogName,
	}

	cmd := exec.CommandContext(ctx, "mysqlbinlog", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func DirUsage(dir string) (int64, error) {
	var usage int64
	fn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}

		st, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			usage += info.Size()
		} else {
			usage += st.Blocks * 512
		}
		return nil
	}
	if err := filepath.WalkDir(dir, fn); err != nil {
		return 0, err
	}

	return usage, nil
}

// SortBinlogs sort binlog filenames according to its number.
// `binlogs` should contains filenames such as `binlog.000001`.
func SortBinlogs(binlogs []string) {
	sort.Slice(binlogs, func(i, j int) bool {
		log1 := binlogs[i]
		log2 := binlogs[j]
		var index1, index2 int64
		if fields := strings.Split(log1, "."); len(fields) == 2 {
			index1, _ = strconv.ParseInt(fields[1], 10, 64)
		}
		if fields := strings.Split(log2, "."); len(fields) == 2 {
			index2, _ = strconv.ParseInt(fields[1], 10, 64)
		}
		return index1 < index2
	})
}

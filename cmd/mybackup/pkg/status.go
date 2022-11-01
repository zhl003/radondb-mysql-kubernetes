package pkg

import (
	"context"
	"fmt"
)

func (o operator) GetServerStatus(ctx context.Context, st *ServerStatus) error {
	ms := &showMasterStatus{}
	if rows, err := o.db.QueryContext(ctx, `SHOW MASTER STATUS`); err != nil {
		return fmt.Errorf("failed to show master status: %w", err)
	} else {
		for rows.Next() {
			if err := rows.Scan(&ms.File, &ms.Position, &ms.BinlogDoDB, &ms.BinlogIgnoreDB, &ms.ExecutedGTIDSet); err != nil {
				return fmt.Errorf("failed to scan master status: %w", err)
			}
		}

	}

	if rows, err := o.db.QueryContext(ctx, `SELECT @@super_read_only, @@server_uuid`); err != nil {
		return fmt.Errorf("failed to get global variables: %w", err)
	} else {
		for rows.Next() {
			if err := rows.Scan(&st.SuperReadOnly, &st.UUID); err != nil {
				return fmt.Errorf("failed to scan global variables: %w", err)
			}
		}
	}

	st.CurrentBinlog = ms.File
	return nil
}

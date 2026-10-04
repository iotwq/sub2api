package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestOpsErrorInsertColumnsPlaceholdersAndArgumentsMatch(t *testing.T) {
	parts := strings.Split(insertOpsErrorLogSQL, ") VALUES (")
	require.Len(t, parts, 2)
	columns := strings.Split(parts[0][strings.Index(parts[0], "(")+1:], ",")
	placeholders := strings.Split(strings.TrimSuffix(strings.TrimSpace(parts[1]), ")"), ",")
	args := opsInsertErrorLogArgs(&service.OpsInsertErrorLogInput{})
	require.Len(t, columns, len(args))
	require.Len(t, placeholders, len(args))
	for i, value := range placeholders {
		require.Equal(t, fmt.Sprintf("$%d", i+1), strings.TrimSpace(value))
	}
}

func TestOpsErrorInsertOmitsRemovedReplayFields(t *testing.T) {
	for _, column := range []string{"request_body", "request_body_truncated", "request_body_bytes"} {
		require.NotContains(t, insertOpsErrorLogSQL, column)
	}
	createdAt := time.Date(2026, 9, 27, 0, 0, 0, 0, time.UTC)
	elapsed := int64(24477)
	legacyBody, legacyBytes := "legacy replay body", 1234
	args := opsInsertErrorLogArgs(&service.OpsInsertErrorLogInput{
		RequestBodyJSON:      &legacyBody,
		RequestBodyTruncated: true,
		RequestBodyBytes:     &legacyBytes,
		CreatedAt:            createdAt,
		APIKeyPrefix:         "sk-ops-regression",
		DurationMs:           &elapsed,
	})
	require.Len(t, args, 39)
	require.Equal(t, createdAt, args[36])
	require.Equal(t, sql.NullString{String: "sk-ops-regression", Valid: true}, args[37])
	require.Equal(t, sql.NullInt64{Int64: elapsed, Valid: true}, args[38])
}

func TestOpsErrorDurationPersistsInSingleAndBatchInserts(t *testing.T) {
	zero, elapsed, invalid := int64(0), int64(24477), int64(-1)
	for _, tc := range []struct {
		name     string
		value    *int64
		expected driver.Value
	}{
		{"missing", nil, nil}, {"zero", &zero, int64(0)}, {"measured", &elapsed, int64(24477)}, {"invalid", &invalid, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newSQLMock(t)
			repo := NewOpsRepository(db)
			input := &service.OpsInsertErrorLogInput{ErrorPhase: "request", ErrorType: "client_canceled", StatusCode: 499, DurationMs: tc.value, CreatedAt: time.Now()}
			args := make([]driver.Value, 39)
			for i := 0; i < 38; i++ {
				args[i] = sqlmock.AnyArg()
			}
			args[38] = tc.expected
			mock.ExpectQuery("(?s)INSERT INTO ops_error_logs.*duration_ms.*RETURNING id").WithArgs(args...).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			id, err := repo.InsertErrorLog(context.Background(), input)
			require.NoError(t, err)
			require.EqualValues(t, 1, id)
			mock.ExpectBegin()
			prepared := mock.ExpectPrepare("(?s)INSERT INTO ops_error_logs.*duration_ms")
			prepared.ExpectExec().WithArgs(args...).WillReturnResult(sqlmock.NewResult(0, 1))
			prepared.ExpectExec().WithArgs(args...).WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()
			n, err := repo.BatchInsertErrorLogs(context.Background(), []*service.OpsInsertErrorLogInput{input, input})
			require.NoError(t, err)
			require.EqualValues(t, 2, n)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

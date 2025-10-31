package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Lysoul/gocommon/monitoring"
	"github.com/alexlast/bunzap"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/uptrace/bun/extra/bunotel"
)

func TestLog(t *testing.T) {
	t.Setenv("LOG_MODE", "dev")
	t.Setenv("LOG_LEVEL", "DEBUG")
	t.Setenv("LOG_ENCODING", "console")
	t.Setenv("SERVICE_NAME", "foo")

	log := monitoring.Logger()

	sqldb, mock, err := sqlmock.New()
	require.NoError(t, err)
	opts := []bun.DBOption{}

	db := bun.NewDB(sqldb, pgdialect.New(), opts...)

	db.AddQueryHook(bunzap.NewQueryHook(bunzap.QueryHookOptions{
		Logger:       log.Logger,
		SlowDuration: 10 * time.Millisecond, // Omit to log all operations as debug
	}))
	db.AddQueryHook(bunotel.NewQueryHook(bunotel.WithDBName("mock")))

	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
		bundebug.WithEnabled(true),
	))

	log.Info("Connected to postgres")

	mock.ExpectQuery("SELECT 1").
		WillDelayFor(100 * time.Millisecond).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	ctx := context.Background()
	res := 0
	err = db.NewRaw("SELECT 1").Scan(ctx, &res)
	require.NoError(t, err)

}

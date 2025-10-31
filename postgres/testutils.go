package postgres

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/extra/bundebug"
)

// Helper for mocking db on other tests
func ConnectMock(t *testing.T) (*bun.DB, sqlmock.Sqlmock) {

	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	db := bun.NewDB(sqldb, pgdialect.New())
	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	return db, mock

}

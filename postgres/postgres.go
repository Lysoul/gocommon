package postgres

import (
	"context"
	"database/sql"
	"net/url"
	"regexp"
	"runtime"
	"time"

	"github.com/Lysoul/gocommon/monitoring"
	"go.uber.org/zap"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/uptrace/bun/extra/bunotel"
	"github.com/uptrace/bun/migrate"
)

// For productopnm config, see
//   - see https://bun.uptrace.dev/guide/running-bun-in-production.html
//   - see https://bun.uptrace.dev/postgres/#pgdriver
type Config struct {
	URL     string `envconfig:"POSTGRES_URL" required:"true"`
	Migrate bool   `envconfig:"POSTGRES_MIGRATE"`
	Debug   bool   `envconfig:"POSTGRES_DEBUG" default:"true"`
	// default to 4 * runtime.NumCPU
	MaxOpenConns int `envconfig:"POSTGRES_MAX_OPEN_CONNS"`
	// default to 4 * runtime.NumCPU
	MaxIdleConns int `envconfig:"POSTGRES_MAX_IDLE_CONNS"`
	// To make your app more resilient to errors during migrations, you can tweak Bun to discard unknown columns in production
	DiscardUnknownColumns bool          `envconfig:"POSTGRES_BUN_DISCARD_UNKNOWN_COLUMNS"`
	SlowQueriesDuration   time.Duration `envconfig:"POSTGRES_SLOW_QUERIES_DURATION" default:"200ms"`
}

func Connect(config Config, migrations *migrate.Migrations) *bun.DB {
	log := monitoring.Logger()
	log.Info("Connecting to database " +
		regexp.MustCompile(`postgres://.+@`).
			ReplaceAllString(config.URL, "postgres://xxxx:xxxx@"))

	if config.MaxOpenConns == 0 && config.MaxIdleConns == 0 {
		maxOpenConns := 4 * runtime.GOMAXPROCS(0)
		config.MaxOpenConns = maxOpenConns
		config.MaxIdleConns = maxOpenConns
	}

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(config.URL)))
	sqldb.SetMaxOpenConns(config.MaxOpenConns)
	sqldb.SetMaxIdleConns(config.MaxIdleConns)

	opts := []bun.DBOption{}
	if config.DiscardUnknownColumns {
		opts = append(opts, bun.WithDiscardUnknownColumns())
	}

	db := bun.NewDB(sqldb, pgdialect.New(), opts...)

	db.AddQueryHook(NewQueryHook(QueryHookOptions{
		Logger:          log.Logger,
		SlowDuration:    config.SlowQueriesDuration, // Omit to log all operations as debug
		IgnoreErrNoRows: true,
	}))
	u, _ := url.Parse(config.URL)
	db.AddQueryHook(bunotel.NewQueryHook(bunotel.WithDBName(u.Path[1:])))

	if config.Debug {
		db.AddQueryHook(bundebug.NewQueryHook(
			bundebug.WithVerbose(true),
			bundebug.WithEnabled(true),
		))
	}
	log.Info("Connected to postgres")

	if config.Migrate && migrations != nil {
		migrator := migrate.NewMigrator(db, migrations)
		ctx := context.Background()
		migrator.Init(ctx)
		group, err := migrator.Migrate(ctx)
		if err != nil {
			log.Error("failed to migrate", zap.Error(err))
			migrator.Rollback(ctx)
			log.Fatal("failed to migrate (but rolled back failed migration)")
		}
		log.Info("migrated to " + group.String())
	}

	monitoring.AddCheck("postgres", false, 3*time.Second, HealthFunc(db))

	return db

}

func HealthFunc(db *bun.DB) func(context.Context) error {
	return func(ctx context.Context) error {
		return db.PingContext(ctx)
	}
}

package postgres

import (
	"strings"

	"github.com/Lysoul/gocommon/monitoring"

	"github.com/kelseyhightower/envconfig"
	"github.com/uptrace/bun/migrate"
	"github.com/urfave/cli/v2"
)

var config = &Config{}

func LoadConfig() {
	envconfig.MustProcess("", config)
}

func CliCommand(migrations *migrate.Migrations) *cli.Command {
	LoadConfig()
	cf := *config

	return &cli.Command{
		Name:  "db",
		Usage: "database migrations",
		Subcommands: []*cli.Command{
			{
				Name:  "init",
				Usage: "create migration tables",
				Action: func(c *cli.Context) error {
					db := Connect(cf, nil)
					migrator := migrate.NewMigrator(db, migrations)
					monitoring.Logger().Info("Initializing bun tables")
					return migrator.Init(c.Context)
				},
			},
			{
				Name:  "migrate",
				Usage: "migrate database",
				Action: func(c *cli.Context) error {
					// connect will auto migrate
					Connect(cf, migrations)
					return nil
				},
			},
			{
				Name:  "rollback",
				Usage: "rollback the last migration group",
				Action: func(c *cli.Context) error {
					log := monitoring.Logger().Sugar()
					db := Connect(cf, nil)
					migrator := migrate.NewMigrator(db, migrations)
					group, err := migrator.Rollback(c.Context)
					if err != nil {
						return err
					}
					if group.IsZero() {
						log.Infof("there are no groups to roll back\n")
						return nil
					}
					log.Infof("rolled back %s\n", group)
					return nil
				},
			},
			{
				Name:  "lock",
				Usage: "lock migrations",
				Action: func(c *cli.Context) error {
					db := Connect(cf, migrations)
					migrator := migrate.NewMigrator(db, migrations)
					return migrator.Lock(c.Context)
				},
			},
			{
				Name:  "unlock",
				Usage: "unlock migrations",
				Action: func(c *cli.Context) error {
					db := Connect(cf, migrations)
					migrator := migrate.NewMigrator(db, migrations)
					return migrator.Unlock(c.Context)
				},
			},
			{
				Name:  "create_go",
				Usage: "create Go migration",
				Action: func(c *cli.Context) error {
					log := monitoring.Logger().Sugar()

					db := Connect(cf, migrations)
					migrator := migrate.NewMigrator(db, migrations)
					name := strings.Join(c.Args().Slice(), "_")
					mf, err := migrator.CreateGoMigration(c.Context, name)
					if err != nil {
						return err
					}
					log.Infof("created migration %s (%s)\n", mf.Name, mf.Path)
					return nil
				},
			},
			{
				Name:  "create-sql",
				Usage: "create up and down SQL migrations",
				Action: func(c *cli.Context) error {
					log := monitoring.Logger().Sugar()
					db := Connect(cf, nil)
					migrator := migrate.NewMigrator(db, migrations)
					name := strings.Join(c.Args().Slice(), "_")
					files, err := migrator.CreateSQLMigrations(c.Context, name)
					if err != nil {
						return err
					}

					for _, mf := range files {
						log.Infof("created migration %s (%s)\n", mf.Name, mf.Path)
					}
					return nil
				},
			},
			{
				Name:  "status",
				Usage: "print migrations status",
				Action: func(c *cli.Context) error {
					log := monitoring.Logger().Sugar()

					db := Connect(cf, nil)
					migrator := migrate.NewMigrator(db, migrations)
					ms, err := migrator.MigrationsWithStatus(c.Context)
					if err != nil {
						return err
					}
					log.Infof("migrations: %s\n", ms)
					log.Infof("unapplied migrations: %s\n", ms.Unapplied())
					log.Infof("last migration group: %s\n", ms.LastGroup())
					return nil
				},
			},
		},
	}
}

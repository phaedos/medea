package migrate

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"medea/pkg/config"
	"medea/pkg/database"
	"medea/pkg/database/migrate"

	"github.com/gookit/color"
	"github.com/jinzhu/gorm"
	"gopkg.in/urfave/cli.v2"

	_ "medea/pkg/database/migrate/migrations"
)

var (
	connection  *gorm.DB
	err         error
	cmdCategory = "migrate"
	migrateTpl  = `
package migrations

import (
	"medea/pkg/database/migrate"
	"github.com/jinzhu/gorm"
)

func init() {
	migrate.DefaultMC.Register(&{{ .StructureName }}{})
}

// {{ .StructureName }} represent some database operate
type {{ .StructureName }} struct{}

// Name represent operate name, it's unique
func (c *{{ .StructureName }}) Name() string {
	return "{{ .Name }}"
}

// Up is executed in upgrading 
func (c *{{ .StructureName }}) Up(db *gorm.DB) error {
	// execute when upgrade database
	return nil
}

// Down is executed in downgrading
func (c *{{ .StructureName }}) Down(db *gorm.DB) error {
	// execute when rollback database
	return nil
}
`
	Commands = []*cli.Command{
		{
			Name:      "migrate:create",
			Category:  cmdCategory,
			Usage:     "create migration files",
			UsageText: "migrate:create [command options] migration_file_name",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "path",
					Aliases: []string{"p"},
					Usage:   "migration files path",
					Value:   "databases/migrate/migrations",
				},
			},
			Action: func(context *cli.Context) error {
				if context.NArg() < 1 {
					return errors.New("migration file name is empty")
				}
				var (
					path          = context.String("path")
					stat          os.FileInfo
					fileName      = context.Args().Get(0)
					structureName string
					timestamp     string
					tpl           *template.Template
					variables     struct {
						StructureName string
						Name          string
					}
					migrationFile *os.File
				)

				if stat, err = os.Stat(path); err != nil {
					return err
				} else if !stat.IsDir() {
					return fmt.Errorf("path must be a directory")
				}

				timestamp = time.Now().Format("20060102150405")
				structureName = strings.Replace(fileName, "_", " ", -1)
				structureName = strings.Title(structureName)
				structureName = strings.Replace(structureName, " ", "", -1)
				variables.StructureName = fmt.Sprintf("%s%s", structureName, timestamp)
				variables.Name = fmt.Sprintf("%s_%s", fileName, timestamp)

				if tpl, err = template.New("migrateBak tpl").Parse(migrateTpl); err != nil {
					return err
				}

				fileName = fmt.Sprintf("%s/%s_%s.go", strings.TrimRight(path, "/"), timestamp, fileName)
				if migrationFile, err = os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0666); err != nil {
					return err
				}
				defer migrationFile.Close()

				if err = tpl.Execute(migrationFile, variables); err != nil {
					return err
				}

				color.Green.Printf("Create migration file: %s\n", fileName)

				return nil
			},
		},
		{
			Name:      "migrate:upgrade",
			Usage:     "execute migrate to upgrade database",
			UsageText: "migrate:upgrade",
			Category:  cmdCategory,
			Action: func(context *cli.Context) error {
				migrate.DefaultMC.Upgrade()
				color.Green.Println("Migrate: done!")
				return nil
			},
			Before: func(context *cli.Context) error {
				connection, err = database.NewConnection(&config.DefaultConfig.Database)
				migrate.DefaultMC.SetConnection(connection)
				return err
			},
		},
		{
			Name:      "migrate:rollback",
			Usage:     "rollback database",
			UsageText: "migrate:rollback [command options]",
			Category:  cmdCategory,
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:    "step",
					Aliases: []string{"s"},
					Usage:   "rollback step",
					Value:   1,
				},
			},
			Action: func(context *cli.Context) error {
				migrate.DefaultMC.Rollback(uint(context.Int("step")))
				color.Red.Println("Rollback: done!")
				return nil
			},
			Before: func(context *cli.Context) error {
				connection, err = database.NewConnection(&config.DefaultConfig.Database)
				migrate.DefaultMC.SetConnection(connection)
				return err
			},
		},
		{
			Name:      "migrate:refresh",
			Usage:     "refresh database",
			UsageText: "migrate:refresh",
			Category:  cmdCategory,
			Action: func(context *cli.Context) error {
				step := migrate.DefaultMC.MaxBatch()
				migrate.DefaultMC.Rollback(step)
				migrate.DefaultMC.Upgrade()
				color.Blue.Println("Refresh: done!")
				return nil
			},
			Before: func(context *cli.Context) error {
				connection, err = database.NewConnection(&config.DefaultConfig.Database)
				migrate.DefaultMC.SetConnection(connection)
				return err
			},
		},
	}
)

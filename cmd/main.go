package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	"medea/pkg/config"
	cmdApp "medea/serve/app"
	"medea/serve/client"
	"medea/serve/http"
	"medea/serve/migrate"

	"medea/pkg/log"

	"github.com/gookit/color"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"gopkg.in/urfave/cli.v2"
)

var (
	app = cli.App{
		Name:      "medea",
		Version:   "0.0.1",
		Compiled:  time.Now(),
		Usage:     "develop toolkit and program entry",
		UsageText: "medea [global options] command [command options] [arguments...]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Usage:   "system config file, search path: medea.yaml, $HOME/medea.yaml, /etc/medea/medea.yaml",
				Aliases: []string{"c"},
				EnvVars: []string{"MEDEA_CONFIG"},
			},
			&cli.StringFlag{
				Name:  "db-host",
				Usage: "set the database host",
				Value: "127.0.0.1",
			},
			&cli.UintFlag{
				Name:  "db-port",
				Usage: "set the database port",
				Value: 3306,
			},
			&cli.StringFlag{
				Name:  "db-user",
				Usage: "set the database user",
				Value: "root",
			},
			&cli.StringFlag{
				Name:  "db-pwd",
				Usage: "set the database password",
				Value: "character",
			},
			&cli.StringFlag{
				Name:  "db-name",
				Usage: "set the database name",
				Value: "medeadb",
			},
		},
		Before: func(ctx *cli.Context) error {
			var (
				err      error
				cfgFile  = ctx.String("config")
				userHome string
			)

			if cfgFile != "" {
				viper.SetConfigFile(cfgFile)
			} else {
				viper.SetConfigName("medea")
				viper.SetConfigType("yaml")
				viper.AddConfigPath(".")

				if userHome, err = homedir.Dir(); err != nil {
					return err
				}

				viper.AddConfigPath(userHome)
				viper.AddConfigPath("/etc/medea")
			}

			if err = viper.ReadInConfig(); err == nil {
				if err = viper.Unmarshal(config.DefaultConfig); err != nil {
					return err
				}
				if _, err = log.NewLogger(&config.DefaultConfig.Log); err != nil {
					return err
				}
			} else {
				color.Warn.Println(err.Error())
				config.DefaultConfig.Database.Host = ctx.String("db-host")
				config.DefaultConfig.Database.Port = uint32(ctx.Uint("db-port"))
				config.DefaultConfig.Database.User = ctx.String("db-user")
				config.DefaultConfig.Database.Password = ctx.String("db-pwd")
				config.DefaultConfig.Database.DBName = ctx.String("db-name")
			}

			return nil
		},
	}
)

func main() {
	var commands []*cli.Command

	commands = append(commands, migrate.Commands...)
	commands = append(commands, cmdApp.Commands...)
	commands = append(commands, client.Commands...)
	commands = append(commands, http.Commands...)
	app.Commands = commands

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}

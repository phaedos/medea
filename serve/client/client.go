package client

import (
	"fmt"
	"medea/pkg/log"

	"gopkg.in/urfave/cli.v2"

	_ "medea/pkg/database/migrate/migrations"
)

var (
	category = "client"
	logger   = log.MustNewLogger(nil)

	Commands = []*cli.Command{
		{
			Name:      "client:token",
			Category:  category,
			Usage:     "client token",
			UsageText: "client:token",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "uid",
					Usage: "app uid",
				},
				&cli.StringFlag{
					Name:  "secret",
					Usage: "app secret",
				},
				&cli.StringFlag{
					Name:  "host",
					Usage: "app host allow",
				},
			},
			Action: func(context *cli.Context) error {
				val := map[string]string{
					"uid":    context.String("uid"),
					"secret": context.String("secret"),
					"host":   context.String("host"),
				}
				if len(val["uid"]) == 0 {
					logger.Error("app uid is empty \n")
				}
				if len(val["secret"]) == 0 {
					logger.Error("app secret is empty \n")
				}

				globalEnvironmentUpdate()
				if len(val["host"]) == 0 {
					val["host"] = medeaHost
				}

				if err := token_create(val); err != nil {
					fmt.Println("token create failed", err)
				}
				return nil
			},
		},
		{
			Name:      "client:file",
			Category:  category,
			Usage:     "client file",
			UsageText: "client:file",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "token",
					Usage: "token",
				},
				&cli.StringFlag{
					Name:  "secret",
					Usage: "secret",
				},
				&cli.StringFlag{
					Name:  "path",
					Usage: "file upload path",
				},
				&cli.StringFlag{
					Name:  "src",
					Usage: "file src path",
				},
				&cli.StringFlag{
					Name:  "host",
					Usage: "app host allow",
				},
			},
			Action: func(context *cli.Context) error {
				val := map[string]string{
					"token":  context.String("token"),
					"secret": context.String("secret"),
					"path":   context.String("path"),
					"src":    context.String("src"),
					"host":   context.String("host"),
				}
				if len(val["token"]) == 0 {
					logger.Error("access token is empty \n")
				}
				if len(val["secret"]) == 0 {
					logger.Error("access secret is empty \n")
				}
				if len(val["secret"]) == 0 {
					logger.Error("app secret is empty \n")
				}
				if len(val["path"]) == 0 {
					logger.Error("upload path is empty \n")
				}
				if len(val["src"]) == 0 {
					logger.Error("src file is empty \n")
				}

				globalEnvironmentUpdate()
				if len(val["host"]) == 0 {
					val["host"] = medeaHost
				}

				if err := file_create(val); err != nil {
					fmt.Println("file create failed", err)
				}
				return nil
			},
		},
		{
			Name:      "client:read",
			Category:  category,
			Usage:     "client read",
			UsageText: "client:read",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "token",
					Usage: "access token",
				},
				&cli.StringFlag{
					Name:  "secret",
					Usage: "access secret",
				},
				&cli.StringFlag{
					Name:  "uid",
					Usage: "file uid",
				},
				&cli.StringFlag{
					Name:  "dst",
					Usage: "dst file name",
				},
				&cli.StringFlag{
					Name:  "host",
					Usage: "app host allow",
				},
				&cli.StringFlag{
					Name:  "range",
					Usage: "range concurrence",
				},
			},
			Action: func(context *cli.Context) error {
				val := map[string]string{
					"token":  context.String("token"),
					"secret": context.String("secret"),
					"uid":    context.String("uid"),
					"dst":    context.String("dst"),
					"host":   context.String("host"),
					"range":  context.String("range"),
				}
				if len(val["token"]) == 0 {
					logger.Error("app token is empty \n")
				}
				if len(val["secret"]) == 0 {
					logger.Error("app secret is empty \n")
				}
				if len(val["uid"]) == 0 {
					logger.Error("file uid is empty \n")
				}
				if len(val["dst"]) == 0 {
					logger.Error("local dst is empty \n")
				}

				globalEnvironmentUpdate()
				if len(val["host"]) == 0 {
					val["host"] = medeaHost
				}

				if err := file_read(val); err != nil {
					fmt.Println("file read failed", err)
				}
				return nil
			},
		},
		{
			Name:      "client:info",
			Category:  category,
			Usage:     "client info",
			UsageText: "client:info",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "token",
					Usage: "access token",
				},
				&cli.StringFlag{
					Name:  "secret",
					Usage: "access secret",
				},
				&cli.StringFlag{
					Name:  "uid",
					Usage: "file uid",
				},
				&cli.StringFlag{
					Name:  "host",
					Usage: "app host allow",
				},
			},
			Action: func(context *cli.Context) error {
				val := map[string]string{
					"token":  context.String("token"),
					"secret": context.String("secret"),
					"uid":    context.String("uid"),
					"host":   context.String("host"),
				}
				if len(val["token"]) == 0 {
					logger.Error("app token is empty \n")
				}
				if len(val["secret"]) == 0 {
					logger.Error("app secret is empty \n")
				}
				if len(val["uid"]) == 0 {
					logger.Error("file uid is empty \n")
				}

				globalEnvironmentUpdate()
				if len(val["host"]) == 0 {
					val["host"] = medeaHost
				}

				if err := file_info(val); err != nil {
					fmt.Println("file info failed", err)
				}
				return nil
			},
		},
		{
			Name:      "client:list",
			Category:  category,
			Usage:     "client list",
			UsageText: "client:list",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "token",
					Usage: "access token",
				},
				&cli.StringFlag{
					Name:  "secret",
					Usage: "access secret",
				},
				&cli.StringFlag{
					Name:  "host",
					Usage: "app host allow",
				},
			},
			Action: func(context *cli.Context) error {
				val := map[string]string{
					"token":  context.String("token"),
					"secret": context.String("secret"),
					"host":   context.String("host"),
				}
				if len(val["token"]) == 0 {
					logger.Error("app token is empty \n")
				}
				if len(val["secret"]) == 0 {
					logger.Error("app secret is empty \n")
				}

				globalEnvironmentUpdate()
				if len(val["host"]) == 0 {
					val["host"] = medeaHost
				}

				if err := file_list(val); err != nil {
					fmt.Println("file list failed", err)
				}
				return nil
			},
		},
		{
			Name:      "client:env",
			Category:  category,
			Usage:     "client env",
			UsageText: "client:env",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "server",
					Usage: "server address",
				},
				&cli.StringFlag{
					Name:  "host",
					Usage: "app host allow",
				},
			},
			Action: func(context *cli.Context) error {
				val := map[string]string{
					"server": context.String("server"),
					"host":   context.String("host"),
				}

				if err := env_update(val); err != nil {
					fmt.Println("env update failed", err)
				}
				return nil
			},
		},
	}
)

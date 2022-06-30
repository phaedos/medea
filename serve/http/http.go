package http

import (
	ctx "context"
	"fmt"
	libHTTP "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"medea/pkg/config"
	"medea/pkg/database"
	"medea/pkg/database/migrate"
	"medea/pkg/http"
	"medea/pkg/log"

	"github.com/gin-gonic/gin"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/urfave/cli.v2"

	_ "medea/pkg/database/migrate/migrations"
)

var (
	category = "http"
	logger   = log.MustNewLogger(nil)

	Commands = []*cli.Command{
		{
			Name:      "http:routes",
			Category:  category,
			Usage:     "list http routes",
			UsageText: "http:routes",
			Action: func(context *cli.Context) error {
				gin.SetMode(gin.ReleaseMode)
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"method", "path", "handler"})
				for _, route := range http.Routers().Routes() {
					table.Append([]string{route.Method, route.Path, route.Handler})
				}
				table.Render()
				return nil
			},
		},
		{
			Name:      "http:start",
			Category:  category,
			Usage:     "start http service",
			UsageText: "http:start [command options]",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "host",
					Aliases: []string{"H"},
					Usage:   "http service listen ip",
					Value:   "0.0.0.0",
				},
				&cli.Int64Flag{
					Name:    "port",
					Aliases: []string{"P"},
					Usage:   "http service listen port",
					Value:   8630,
				},
				&cli.DurationFlag{
					Name: "read-timeout",
					Usage: "read-timeout is the maximum duration for reading " +
						"the entire request, including the body",
					Value: 0,
				},
				&cli.DurationFlag{
					Name: "read-header-timeout",
					Usage: "read-header-timeout is the amount of time allowed " +
						"to read request headers",
					Value: 0,
				},
				&cli.DurationFlag{
					Name: "write-timeout",
					Usage: "writer-timeout is the maximum duration before timing " +
						"out writes of the response",
					Value: 0,
				},
				&cli.DurationFlag{
					Name: "idle-timeout",
					Usage: "idle-timeout is the maximum amount of time to wait for " +
						"the next request when keep-alives are enabled",
					Value: 0,
				},
				&cli.DurationFlag{
					Name:  "wait-shutdown",
					Usage: "wait time before timeout for closing server",
					Value: 5 * time.Second,
				},
				&cli.IntFlag{
					Name: "max-header-bytes",
					Usage: "max-header-bytes controls the maximum number of bytes the " +
						"server will read parsing the request header's keys and values, " +
						"including the request line. It does not limit the size of the request body",
					Value: 0,
				},
				&cli.StringFlag{
					Name:  "cert-file",
					Usage: "certificate file for starting https service",
				},
				&cli.StringFlag{
					Name:  "cert-key",
					Usage: "certificate key file for starting https service",
				},
			},
			Action: func(context *cli.Context) error {
				addr := fmt.Sprintf("%s:%d", context.String("host"), context.Int64("port"))
				server := libHTTP.Server{
					Addr:              addr,
					Handler:           http.Routers(),
					ReadTimeout:       context.Duration("read-timeout"),
					ReadHeaderTimeout: context.Duration("read-header-timeout"),
					WriteTimeout:      context.Duration("write-timeout"),
					IdleTimeout:       context.Duration("idle-timeout"),
					MaxHeaderBytes:    context.Int("max-header-bytes"),
				}
				certFile := context.String("cert-file")
				certKey := context.String("cert-key")

				go func() {
					if certFile != "" && certKey != "" {
						logger.Infof("medea https service listening on: https://%s", addr)
						if err := server.ListenAndServeTLS(certFile, certKey); err != nil && err != libHTTP.ErrServerClosed {
							logger.Errorf("https server error: %s", err)
						}
					} else {
						logger.Infof("medea http service listening on: http://%s", addr)
						if err := server.ListenAndServe(); err != nil && err != libHTTP.ErrServerClosed {
							logger.Errorf("https server error: %s", err)
						}

					}

				}()

				quit := make(chan os.Signal, 1)
				signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
				<-quit
				logger.Debug("Shutdown Server ...")

				ctx, cancel := ctx.WithTimeout(ctx.Background(), context.Duration("wait-shutdown"))
				defer cancel()
				if err := server.Shutdown(ctx); err != nil {
					logger.Fatal("Server Shutdown:", err)
				}
				<-ctx.Done()
				logger.Debugf("Shutdown timeout of %s", context.Duration("wait-shutdown"))
				logger.Debug("Server exiting")
				return nil
			},
			Before: func(context *cli.Context) (err error) {
				gin.SetMode(gin.ReleaseMode)
				db := database.MustNewConnection(&config.DefaultConfig.Database)
				migrate.DefaultMC.SetConnection(db)
				migrate.DefaultMC.Upgrade()
				return nil
			},
		},
	}
)

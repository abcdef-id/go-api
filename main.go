package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/abcdef-id/go-api/logger"
	"github.com/abcdef-id/go-api/routes"
	"github.com/labstack/echo"

	"github.com/labstack/echo/middleware"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "",
		Short: "Start Service",

		Run: func(cmd *cobra.Command, args []string) {

			e := echo.New()
			e.Pre(middleware.RemoveTrailingSlash())
			e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
				Output: logger.APILog,
			}))
			e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
				AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
			}))
			// Handler for hooking any request in routers registered and log it
			e.Use(middleware.BodyDumpWithConfig(middleware.BodyDumpConfig{
				Handler: logger.APILogHandler,
				Skipper: logger.APILogSkipper,
			}))
			// Handler for putting teak request and response timestamp. This used for get elapsed time
			e.Use(ServiceRequestTime)

			routes.Api(e)

			// Start server
			// e.Logger.Fatal(e.Start(":8000"))
			go func() {
				if err := e.Start(":" + viper.GetString("port")); err != nil {
					e.Logger.Info("Shutting down the server")
				}
			}()

			// Wait for interrupt signal to gracefully shutdown the server with
			// a timeout of 10 seconds.
			quit := make(chan os.Signal)
			signal.Notify(quit, os.Interrupt)
			<-quit
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := e.Shutdown(ctx); err != nil {
				e.Logger.Fatal(err)
			}

		},
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// ServiceRequestTime middleware adds a `Server` header to the response.
func ServiceRequestTime(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Request().Header.Set("X-RequestTime", time.Now().Format(time.RFC3339))
		return next(c)
	}
}

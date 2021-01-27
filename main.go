//go:generate go run -tags=dev assets_generate.go

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/kamilsamaj/go-vue-app/assets"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

var Version = "development"

func main() {
	var (
		err               error
		nodeContextCancel context.CancelFunc
	)

	logger := logrus.New().WithField("who", "Example")

	/*
	 * Start the Go application
	 */
	httpServer := echo.New()
	httpServer.Use(middleware.CORS())

	httpServer.GET("/*", echo.WrapHandler(http.FileServer(assets.Assets)))
	httpServer.GET("/api/version", func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, Version)
	})

	go func() {
		var err error

		logger.WithFields(logrus.Fields{
			"serverVersion": Version,
		}).Infof("Starting application")

		err = httpServer.Start("0.0.0.0:8080")

		if err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Unable to start application")
		} else {
			logger.Info("Shutting down the server...")
		}
	}()

	/*
	 * Setup shutdown handler
	 */
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM)

	/*
	 * Start the Node client app (only for version "development")
	 */
	if Version == "development" {
		_, nodeContextCancel = context.WithCancel(context.Background())

		go func() {
			logger.Info("Starting Node development server...")

			var cmd *exec.Cmd
			var err error

			if cmd, err = StartClientApp(); err != nil {
				logger.WithError(err).Fatal("Error starting Node development!")
			}

			cmd.Wait()
			logger.Info("Stopping Node development server...")
		}()
	}

	/*
	 * Wait for and stop both the Go and Node apps
	 */
	<-quit

	if Version == "development" {
		nodeContextCancel()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = httpServer.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("There was an error shutting down the server")
	}

	logger.Info("Application stopped")
}

/*
StartClientApp runs your NodeJS client app found in the "app" directory
*/
func StartClientApp() (*exec.Cmd, error) {
	var err error

	cmd := exec.Command("npm", "run", "serve")
	cmd.Dir = "./app"
	cmd.Stdout = os.Stdout

	if err = cmd.Start(); err != nil {
		return cmd, fmt.Errorf("error starting NPM: %w", err)
	}

	return cmd, nil
}

//nolint:errcheck
package cmd

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	ec2 "github.com/bananaops/cloudoff/internal/aws"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/spf13/cobra"
)

var serv = &cobra.Command{
	Use:   "serv",
	Short: "Run cloudoff server",
	Run: func(cmd *cobra.Command, args []string) {

		//define logger for http server error
		handler := slog.NewJSONHandler(os.Stdout, nil)
		httplogger := slog.NewLogLogger(handler, slog.LevelError)

		// Create a new mux for metrics
		muxMetrics := http.NewServeMux()

		// Add a handler for the /metrics endpoint
		muxMetrics.Handle("/metrics", promhttp.Handler())

		metricsServer := &http.Server{
			Addr:              "0.0.0.0:8080",
			ReadHeaderTimeout: 2 * time.Second, // Fix CWE-400 Potential Slowloris Attack because ReadHeaderTimeout is not configured in the http.Server
			Handler:           muxMetrics,
			ErrorLog:          httplogger,
		}

		ec2 := ec2.DiscoverEC2Instances()
		if len(ec2) == 0 {
			slog.Info("no ec2 instances found")
		}

		go func() {
			// Exposer prometheus metrics
			slog.Info("metrics server listening on :8080")
			if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal(fmt.Printf("Failed to serve metrics server: %v\n", err))
				os.Exit(1)
			}
		}()

		// Handle graceful shutdown
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-stop

		slog.Info("shutting down application...")

		// Gracefully stop Metrics server
		if err := metricsServer.Shutdown(context.Background()); err != nil {
			log.Fatal(fmt.Printf("failed to shutdown metrics server: %v\n", err))
		}

		slog.Info("application stopped")

	},
}

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	rootCmd.AddCommand(serv)

}

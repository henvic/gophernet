package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/henvic/gophernet/internal/api"
)

//go:embed data/initial.json
var initialData []byte

func main() {
	var p program
	ctx := context.Background()
	if err := p.run(ctx); err != nil {
		p.log.LogAttrs(ctx, slog.LevelError, err.Error())
		os.Exit(1)
	}
}

// program is used to initialize the GopherNet application.
type program struct {
	log *slog.Logger
	api *api.API
}

// run is a wrapper to handle errors in an easier way.
func (p *program) run(ctx context.Context) (err error) {
	var (
		settings        api.Settings
		initialDataFile string
		reportFile      string
	)

	flag.StringVar(&initialDataFile, "initial-data-file", "", "Initial data (if empty, the server will start with embedded data)")
	flag.StringVar(&reportFile, "report-file", "gophernet-report.txt", "Report output file")
	flag.StringVar(&settings.HTTPAddress, "http", "localhost:8080", "HTTP service address to listen for incoming requests on")
	flag.DurationVar(&settings.UpdateStatusTicker, "update-status-ticker", time.Minute, "Update status ticker interval")
	flag.DurationVar(&settings.BurrowExpiration, "burrow-expiration", 25*24*time.Hour, "The time after which a burrow is considered expired")
	flag.DurationVar(&settings.ReportStatusTicker, "report-status-ticker", 10*time.Minute, "Report status ticker interval")
	flag.Float64Var(&settings.BurrowDigRate, "burrow-dig-rate", 0.9, "The rate at which a gopher digs a burrow (in m/min)")
	flag.BoolVar(&settings.Verbose, "verbose", false, "Show more information about an operation")

	flag.Parse()

	// Set logger.
	var logLevel slog.Leveler
	if settings.Verbose {
		logLevel = slog.LevelDebug
	}
	p.log = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)

	// Load initial burrows.
	var burrows []api.Burrow
	if err := p.loadBurrows(ctx, initialDataFile, &burrows); err != nil {
		return err
	}

	p.api = api.NewAPI(settings, p.log, burrows)
	server := api.NewHTTPServer(p.log, p.api, settings.HTTPAddress)

	// Run the HTTP server.
	ec := make(chan error, 1)
	go func() {
		ec <- server.Run(ctx)
	}()

	// Update the state of the burrows every UpdateStatusTicker.
	go p.api.UpdateStatus(ctx)

	// Report the state of the burrows every ReportStatusTicker.
	go func() {
		// TODO(henvic): handle graceful shutdown gracefully (including closing file to handle flushing).
		if err := p.reportStatusBurrows(ctx, reportFile); err != nil {
			log.Fatal(err)
		}
	}()

	// Waits for the server to shutdown due to graceful termination or otherwise.
	// After a shutdown signal, HTTP requests taking longer than the specified grace period are forcibly closed.
	const grace = time.Second
	select {
	case err = <-ec:
	case <-ctx.Done():
		fmt.Println()
		haltCtx, cancel := context.WithTimeout(context.Background(), grace)
		defer cancel()
		server.Shutdown(haltCtx)
		stop()
		err = <-ec
	}

	if err != nil {
		return fmt.Errorf("application terminated by error: %w", err)
	}
	return nil
}

// loadBurrows loads initial data from a file or embedded data.
func (p *program) loadBurrows(ctx context.Context, filename string, burrows *[]api.Burrow) error {
	var data = initialData
	if filename != "" {
		var err error
		data, err = os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to read initial data file: %w", err)
		}
	}

	if err := json.Unmarshal(data, &burrows); err != nil {
		return fmt.Errorf("failed to unmarshal initial data: %w", err)
	}

	p.log.LogAttrs(ctx, slog.LevelInfo, "Loaded initial data", slog.Int("burrows", len(*burrows)))
	return nil
}

// reportStatusBurrows saves teh state of the burrows to a file, every minute.
func (p *program) reportStatusBurrows(ctx context.Context, reportFile string) (err error) {
	f, err := os.Create(reportFile)
	if err != nil {
		return fmt.Errorf("failed to create report file: %w", err)
	}
	defer func() {
		if err1 := f.Close(); err1 != nil && err == nil {
			err = err1
		}
	}()
	p.api.ReportStatus(ctx, f)
	return nil
}

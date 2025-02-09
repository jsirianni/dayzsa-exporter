// Package main is the entry point of the application.
package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/jsirianni/dayzsa-exporter/client"
	"github.com/jsirianni/dayzsa-exporter/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// CollectionInterval is the default interval
	CollectionInterval = 60 * time.Second
)

var (
	mp = otel.Meter("eventbus/nats")

	playerCount metric.Int64Gauge

	defaultServers = []string{
		// deer isle
		"50.108.13.235:2424",
		// deer isl hardcore
		"50.108.13.235:2324",
		// namaslk hardcore
		"50.108.13.235:2315",
		// frostline
		"50.108.13.235:27016",
	}
)

// DZSA is the DayZ Server Agent
type DZSA struct {
	client   client.Client
	logger   *zap.Logger
	interval time.Duration
	servers  []server
	cancel   context.CancelFunc
}

type server struct {
	ip   string
	port int
}

func main() {
	logger, err := setupLogger()
	if err != nil {
		fmt.Println("Failed to configure logger", err)
		os.Exit(1)
	}

	defaultServers := strings.Join(defaultServers, ",")

	intervalFlag := flag.Duration("collection-interval", CollectionInterval, "The interval at which to collect data from the DayZ server.")
	serversFlag := flag.String("servers", defaultServers, "A comma separated list of DayZ servers to query. In the form of ip:port.")
	flag.Parse()

	if serversFlag == nil {
		logger.Error("servers flag is required")
		os.Exit(1)
	}

	interval := *intervalFlag

	servers := []server{}
	for _, s := range strings.Split(*serversFlag, ",") {
		host, port, err := net.SplitHostPort(s)
		if err != nil {
			logger.Error("invalid server", zap.String("server", s))
			os.Exit(1)
		}

		portNum, err := strconv.Atoi(port)
		if err != nil {
			logger.Error("invalid port", zap.String("port", port))
			os.Exit(1)
		}

		s := server{
			ip:   host,
			port: portNum,
		}
		servers = append(servers, s)
	}

	ctx, cancel := context.WithCancel(context.Background())
	signalCtx, signalCancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer signalCancel()
	c, err := client.New()
	if err != nil {
		logger.Error("new client", zap.Error(err))
		os.Exit(1)
	}

	dzsa := DZSA{
		client:   c,
		logger:   logger,
		interval: interval,
		servers:  servers,
		cancel:   cancel,
	}

	if err := dzsa.setupMetrics(ctx); err != nil {
		logger.Error("setup metrics", zap.Error(err))
		os.Exit(1)
	}

	wg := sync.WaitGroup{}
	for _, s := range dzsa.servers {
		wg.Add(1)
		go func(ip string, port int) {
			defer wg.Done()
			dzsa.watchServer(signalCtx, ip, port)
		}(s.ip, s.port)
	}
	wg.Wait()

	dzsa.logger.Info("Client shutdown complete")
	cancel()
	dzsa.logger.Info("Server stopped")
	os.Exit(0)
}

func (dzsa *DZSA) watchServer(ctx context.Context, ip string, port int) {
	clientName := net.JoinHostPort(ip, strconv.Itoa(port))
	logger := dzsa.logger.With(zap.String("client", clientName))
	logger.Info("starting client")

	for {
		ticker := time.NewTicker(dzsa.interval)
		select {
		case <-ticker.C:
			resp, err := dzsa.client.Query(ip, port)
			if err != nil {
				logger.Error("query", zap.Error(err))
				continue
			}

			playerCount.Record(ctx,
				int64(resp.Result.Players),
				metric.WithAttributeSet(
					attribute.NewSet(
						attribute.String("name", resp.Result.Name),
						attribute.String("endpoint", resp.Result.Endpoint.String()),
					),
				),
			)

			logger.Debug("player count", zap.Int("count", resp.Result.Players))

		case <-ctx.Done():
			logger.Info("shutting down")
			return
		}
	}
}

func (dzsa *DZSA) setupMetrics(ctx context.Context) error {
	dzsa.logger.Info("starting metrics server")

	prometheus, err := metrics.NewPrometheus()
	if err != nil {
		return fmt.Errorf("new prometheus: %w", err)
	}

	if err := prometheus.Start(ctx); err != nil {
		return fmt.Errorf("start prometheus exporter: %w", err)
	}

	playerCount, err = mp.Int64Gauge(
		"dayz.playerCount",
	)
	if err != nil {
		return fmt.Errorf("create player count gauge: %w", err)
	}

	go func() {
		err := httpServer()
		if err != nil {
			dzsa.logger.Error("http server", zap.Error(err))
			dzsa.logger.Error("triggering server shutdown")
			dzsa.cancel()
		}
	}()

	dzsa.logger.Info("metrics server started")

	return nil
}

func httpServer() error {
	s := &http.Server{
		Addr:              "localhost:9100",
		IdleTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	s.Handler = promhttp.Handler()

	return s.ListenAndServe()
}

func setupLogger() (*zap.Logger, error) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.CallerKey = ""
	encoderConfig.StacktraceKey = ""
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.MessageKey = "message"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zap.DebugLevel,
	)

	logger := zap.New(core)
	return logger, nil
}

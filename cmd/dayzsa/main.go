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
	"sync"
	"syscall"
	"time"

	"github.com/jsirianni/dayzsa-exporter/client"
	"github.com/jsirianni/dayzsa-exporter/config"
	"github.com/jsirianni/dayzsa-exporter/internal/ifconfig"
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
	mp = otel.Meter("dayzsa")

	playerCount metric.Int64Gauge
	upStatus    metric.Int64Gauge
)

// DZSA is the DayZ Server Agent
type DZSA struct {
	client   client.Client
	ifconfig ifconfig.Client
	logger   *zap.Logger
	config   *config.Config
	cancel   context.CancelFunc
}

func main() {
	logger, err := setupLogger()
	if err != nil {
		fmt.Println("Failed to configure logger", err)
		os.Exit(1)
	}

	configFile := flag.String("config", "/etc/dayzsa/config.yaml", "The path to the configuration file.")
	flag.Parse()

	conf, err := config.NewFromFile(*configFile)
	if err != nil {
		logger.Error("Failed to create config from file", zap.String("path", *configFile), zap.Error(err))
		os.Exit(1)
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
		ifconfig: ifconfig.New(logger.With(zap.String("module", "ifconfig"))),
		logger:   logger,
		config:   conf,
		cancel:   cancel,
	}

	enableIfConfig := false
	for _, s := range dzsa.config.Servers {
		if s.OverrideIP {
			enableIfConfig = true
			break
		}
	}
	if enableIfConfig {
		dzsa.logger.Info("starting ifconfig dynamic ip detection")
		err := dzsa.ifconfig.Start(ctx)
		if err != nil {
			logger.Error("start ifconfig", zap.Error(err))
			os.Exit(1)
		}
	}

	if err := dzsa.setupMetrics(ctx); err != nil {
		logger.Error("setup metrics", zap.Error(err))
		os.Exit(1)
	}

	wg := sync.WaitGroup{}
	for _, s := range dzsa.config.Servers {
		wg.Add(1)
		go func(s config.Server) {
			defer wg.Done()
			dzsa.watchServer(signalCtx, s)
		}(s)
	}
	wg.Wait()

	dzsa.logger.Info("client shutdown complete")
	cancel()
	dzsa.logger.Info("cerver stopped")
	os.Exit(0)
}

func (dzsa *DZSA) watchServer(ctx context.Context, s config.Server) {
	logger := dzsa.logger.With(
		zap.String("module", "client"),
		zap.String("name", s.Name),
		zap.String("ip", s.IP),
		zap.Int("port", s.Port),
		zap.Bool("override_ip", s.OverrideIP),
	)
	logger.Info("starting client")

	for {
		ticker := time.NewTicker(dzsa.config.Interval)
		select {
		case <-ticker.C:
			ip := s.IP
			port := s.Port

			if s.OverrideIP {
				ipAddr := dzsa.ifconfig.GetAddress()
				if ipAddr != "" {
					logger.Debug("detected public ip for override", zap.String("public_ip", ipAddr))
					ip = ipAddr
				}
			}

			resp, err := dzsa.client.Query(ip, port)
			if err != nil {
				logger.Error("query", zap.Error(err))
				// Record up status as 0 (down) when query fails
				upStatus.Record(ctx, 0,
					metric.WithAttributeSet(
						attribute.NewSet(
							attribute.String("name", s.Name),
							attribute.String("endpoint", net.JoinHostPort(ip, fmt.Sprintf("%d", port))),
						),
					),
				)
				continue
			}

			if resp.Result.Name == "" {
				logger.Error("empty server name")
				// Record up status as 0 (down) when server name is empty
				upStatus.Record(ctx, 0,
					metric.WithAttributeSet(
						attribute.NewSet(
							attribute.String("name", s.Name),
							attribute.String("endpoint", net.JoinHostPort(ip, fmt.Sprintf("%d", port))),
						),
					),
				)
				continue
			}

			// Record up status as 1 (up) when query succeeds
			upStatus.Record(ctx, 1,
				metric.WithAttributeSet(
					attribute.NewSet(
						attribute.String("name", resp.Result.Name),
						attribute.String("endpoint", resp.Result.Endpoint.String()),
					),
				),
			)

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

	upStatus, err = mp.Int64Gauge(
		"dayz.up",
	)
	if err != nil {
		return fmt.Errorf("create up status gauge: %w", err)
	}

	go func() {
		err := dzsa.httpServer()
		if err != nil {
			dzsa.logger.Error("http server", zap.Error(err))
			dzsa.logger.Error("triggering server shutdown")
			dzsa.cancel()
		}
	}()

	dzsa.logger.Info("metrics server started")

	return nil
}

func (dzsa *DZSA) httpServer() error {
	addr := net.JoinHostPort(dzsa.config.Host, "9100")

	s := &http.Server{
		Addr:              addr,
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

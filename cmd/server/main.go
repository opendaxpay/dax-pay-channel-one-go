package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"daxpay.open/dax-pay-channel-one-go/internal/config"
	"daxpay.open/dax-pay-channel-one-go/internal/i18n"
	"daxpay.open/dax-pay-channel-one-go/internal/middleware"
	"daxpay.open/dax-pay-channel-one-go/internal/server"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfgPath := os.Getenv("DAXPAY_CONFIG")
	if cfgPath == "" {
		cfgPath = findConfig()
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		slog.Error("load config failed", "err", err)
		os.Exit(1)
	}

	// 加载 i18n（embed FS 根目录为 i18n/）
	ms, err := i18n.Load(i18n.FS, "i18n")
	if err != nil {
		slog.Error("load i18n failed", "err", err)
		os.Exit(1)
	}
	i18n.SetGlobal(ms)

	ctx := context.Background()
	shutdownTracer, err := middleware.InitTracer(ctx, server.ServiceName, cfg.Tracing.SampleRatio)
	if err != nil {
		slog.Error("init tracer failed", "err", err)
		os.Exit(1)
	}
	defer func() {
		_ = shutdownTracer(context.Background())
	}()

	engine := server.NewRouter()
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("daxpay-channel-one-go starting", "addr", addr, "config", cfgPath)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
	slog.Info("server stopped")
}

func findConfig() string {
	candidates := []string{
		"configs/config.yaml",
		filepath.Join("..", "..", "configs", "config.yaml"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

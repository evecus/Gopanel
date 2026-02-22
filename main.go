package main

import (
	"context"
	"embed"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gopanel/gopanel/internal/api"
	"github.com/gopanel/gopanel/internal/cache"
	"github.com/gopanel/gopanel/internal/config"
	"github.com/gopanel/gopanel/internal/store"
	"github.com/gopanel/gopanel/internal/websocket"
)

//go:embed web/dist
var webFS embed.FS

var version = "dev"

func main() {
	cfgPath := flag.String("config", "config.yaml", "config file path")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Printf("no config file, using defaults: %v", err)
		cfg = config.Default()
	}

	api.SetConfigPath(*cfgPath)

	db, err := store.Init(cfg.DBPath)
	if err != nil {
		log.Fatalf("db init: %v", err)
	}
	defer db.Close()

	hub := websocket.NewHub()
	go hub.Run()
	go store.StartCollector(db, hub, cfg.CollectInterval)

	// 启动服务端缓存，每30秒后台刷新 docker 和 services 数据
	cache.Start(30 * time.Second)
	api.AppVersion = version

	router := api.SetupRouter(cfg, db, hub, webFS)

	srv := &http.Server{
		Addr:         cfg.Listen,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("GoPanel %s running at http://%s", version, cfg.Listen)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

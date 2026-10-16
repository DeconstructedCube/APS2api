package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	_ "time/tzdata"

	"AnonymousVertex/internal/api"
	"AnonymousVertex/internal/buildinfo"
	"AnonymousVertex/internal/config"
	"AnonymousVertex/internal/db"
	"AnonymousVertex/internal/logger"
	"AnonymousVertex/internal/nodes"
	"AnonymousVertex/internal/spool"
	"AnonymousVertex/internal/transport"
	"AnonymousVertex/internal/vertex"
)

const fallbackVersion = "1.3.0-dev"

var (
	//nolint:gochecknoglobals // Injected by build script
	version = fallbackVersion
	//nolint:gochecknoglobals // Injected by build script
	buildCommit = "unknown"
	//nolint:gochecknoglobals // Injected by build script
	buildTime = "unknown"
)

const shutdownGrace = 25 * time.Second

func printBanner() {
	bi := buildinfo.Current()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Printf("║  AnonymousVertex (APS2api)  %-32s ║\n", bi.Version)
	fmt.Println("║  High-performance Anonymous Vertex AI Proxy Server           ║")
	fmt.Printf("║  Build: %-20s  Platform: %s/%-10s     ║\n", bi.Commit, bi.GOOS, bi.GOARCH)
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
}

func main() {
	if version == "" || version == "dev" {
		version = fallbackVersion
	}
	source := "local"
	if buildCommit != "" && buildCommit != "unknown" && buildTime != "" && buildTime != "unknown" {
		source = "release"
	}
	buildinfo.SetInjected(version, buildCommit, buildTime, source)
	setupTermuxCerts() // 优先初始化 Termux 证书

	printBanner()

	logDir := filepath.Join(filepath.Dir(config.ConfigDir()), "logs")
	dailyLogger := logger.NewDailyLogger(logDir)

	cfg := config.GetProvider()
	if err := db.InitDB(filepath.Join(config.ConfigDir(), "data.db")); err != nil {
		log.Fatalf("[APS2api] failed to init database: %v", err)
	}
	defer db.CloseDB()

	if legacyProxy := strings.TrimSpace(cfg.ProxyURL()); legacyProxy != "" {
		if err := transport.ValidateProxyURI(legacyProxy); err != nil {
			log.Printf("[APS2api] 旧 proxy_url 无法构造，保留旧路线: %v", err)
		} else if err := config.MigrateLegacyProxy(legacyProxy); err != nil {
			log.Printf("[APS2api] 迁移旧 proxy_url 到入口代理数据库失败，保留旧路线: %v", err)
		}
	}

	spool.SetMaxSpillBytes(int64(cfg.MaxSpillMB()) << 20)

	nodes.DeleteNodeCallback = transport.RemoveProxy
	transport.StartProxyGC(5*time.Minute, 30*time.Minute)

	keys := api.NewAPIKeyManager()
	keys.LoadKeys()

	api.EnsureAdminPassword()
	stopAdminSessionCleanup := api.StartAdminSessionCleanup(time.Hour)
	defer stopAdminSessionCleanup()

	vc := vertex.NewVertexAIClient(cfg)
	stopEntryProxyProbe := api.StartEntryProxyProbeLoop(vc.Net())
	defer stopEntryProxyProbe()

	srv := api.NewServer(vc, keys, cfg)
	httpServer := &http.Server{
		Addr:              "0.0.0.0:" + strconv.Itoa(cfg.PortAPI()),
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 15 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	shutdownDone := make(chan struct{})
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		defer signal.Stop(sig)
		for {
			s := <-sig
			if s == syscall.SIGHUP {
				config.InvalidateCache()
				config.InvalidateModelsCache()
				log.Printf("[APS2api] 收到 SIGHUP：已清配置/模型缓存，下次读取即热重载")
				continue
			}
			reason := s.String()
			log.Printf("[APS2api] 收到 %s：开始平滑关闭，等待在途请求完成(最长 %s)…", reason, shutdownGrace)
			ctx, cancel := context.WithTimeout(context.Background(), shutdownGrace)
			if err := httpServer.Shutdown(ctx); err != nil {
				log.Printf("[APS2api] 关闭超时/出错：%v(强制结束)", err)
			}
			cancel()
			srv.Close()
			transport.StopAllProxies()
			_ = dailyLogger.Close()
			close(shutdownDone)
			return
		}
	}()

	poolStr := "关闭"
	if cfg.ParallelPoolEnabled() {
		poolStr = strconv.Itoa(cfg.ParallelPoolSize()) + "（开启）"
	}
	log.Printf("[APS2api] 监听 %s（API 密钥 %d 个，max_retries=%d，并发池=%s）",
		httpServer.Addr, keys.Count(), cfg.MaxRetries(), poolStr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[APS2api] server error: %v", err)
	}
	<-shutdownDone
	log.Printf("[APS2api] 关闭完成，程序退出")
}

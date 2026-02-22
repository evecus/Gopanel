package api

import (
	"database/sql"
	"embed"
	"io/fs"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/gopanel/gopanel/internal/api/middleware"
	"github.com/gopanel/gopanel/internal/collector"
	"github.com/gopanel/gopanel/internal/config"
	"github.com/gopanel/gopanel/internal/store"
	ws "github.com/gopanel/gopanel/internal/websocket"
)

var configPath string

func SetConfigPath(p string) { configPath = p }

func SetupRouter(cfg *config.Config, db *sql.DB, hub *ws.Hub, webFS embed.FS) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Authorization", "Content-Type"},
	}))

	api := r.Group("/api")
	api.POST("/login", loginHandler(cfg))
	api.GET("/ws", func(c *gin.Context) { hub.ServeWS(c.Writer, c.Request) })

	auth := api.Group("/")
	auth.Use(middleware.Auth(cfg.JWTSecret))
	{
		auth.GET("/system",      func(c *gin.Context) { c.JSON(200, collector.GetSystemInfo()) })
		auth.GET("/cpu",         func(c *gin.Context) { c.JSON(200, collector.GetCPUStats()) })
		auth.GET("/memory",      func(c *gin.Context) { c.JSON(200, collector.GetMemoryStats()) })
		auth.GET("/disk",        func(c *gin.Context) { c.JSON(200, collector.GetDiskStats()) })
		auth.GET("/network",     func(c *gin.Context) { c.JSON(200, collector.GetNetworkStats()) })
		auth.GET("/temperature", func(c *gin.Context) { c.JSON(200, collector.GetTemperatures()) })
		auth.GET("/crontab",     func(c *gin.Context) { c.JSON(200, collector.GetCrontabs()) })

		auth.GET("/processes", func(c *gin.Context) {
			sortBy := c.DefaultQuery("sort", "cpu")
			sortDir := c.DefaultQuery("dir", "desc")
			limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
			procs, err := collector.GetProcesses(sortBy, sortDir, limit)
			if err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
			c.JSON(200, procs)
		})
		auth.DELETE("/processes/:pid", func(c *gin.Context) {
			pid, err := strconv.ParseInt(c.Param("pid"), 10, 32)
			if err != nil { c.JSON(400, gin.H{"error": "invalid pid"}); return }
			if err := collector.KillProcess(int32(pid)); err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
			c.JSON(200, gin.H{"ok": true})
		})

		auth.GET("/docker/containers", func(c *gin.Context) {
			containers, err := collector.GetContainers()
			if err != nil { c.JSON(200, []interface{}{}); return }
			c.JSON(200, containers)
		})
		auth.POST("/docker/containers/:id/:action", func(c *gin.Context) {
			id, action := c.Param("id"), c.Param("action")
			if !map[string]bool{"start": true, "stop": true, "restart": true}[action] {
				c.JSON(400, gin.H{"error": "invalid action"}); return
			}
			if err := collector.ContainerAction(id, action); err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
			c.JSON(200, gin.H{"ok": true})
		})
		auth.GET("/docker/containers/:id/logs", func(c *gin.Context) {
			lines, _ := strconv.Atoi(c.DefaultQuery("lines", "200"))
			logs, err := collector.GetContainerLogs(c.Param("id"), lines)
			if err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
			c.JSON(200, gin.H{"logs": logs})
		})
		auth.GET("/docker/containers/:id/inspect", func(c *gin.Context) {
			result, err := collector.InspectContainer(c.Param("id"))
			if err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
			c.JSON(200, result)
		})
		auth.POST("/docker/containers/:id/update", func(c *gin.Context) {
			log, err := collector.PullAndUpdateContainer(c.Param("id"))
			if err != nil { c.JSON(500, gin.H{"error": err.Error(), "log": log}); return }
			c.JSON(200, gin.H{"log": log})
		})
		auth.GET("/docker/compose/file", func(c *gin.Context) {
			path := c.Query("path")
			if path == "" { c.JSON(400, gin.H{"error": "path required"}); return }
			content, err := collector.ReadComposeFile(path)
			if err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
			c.JSON(200, gin.H{"content": content})
		})
		auth.POST("/docker/compose/apply", func(c *gin.Context) {
			var req struct {
				Path string `json:"path"`
				Content string `json:"content"`
				ContainerID string `json:"container_id"`
			}
			if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"error": "invalid request"}); return }
			if req.Path == "" { c.JSON(400, gin.H{"error": "path required"}); return }
			log2, err := collector.WriteAndApplyCompose(req.Path, req.Content, req.ContainerID)
			if err != nil { c.JSON(500, gin.H{"error": err.Error(), "log": log2}); return }
			c.JSON(200, gin.H{"message": "重建成功", "log": log2})
		})

		auth.GET("/services", func(c *gin.Context) {
			sortBy  := c.DefaultQuery("sort", "")
			sortDir := c.DefaultQuery("dir", "desc")
			services, err := collector.GetServices()
			if err != nil { c.JSON(200, []interface{}{}); return }
			collector.SortServices(services, sortBy, sortDir)
			c.JSON(200, services)
		})
		auth.POST("/services/:unit/:action", func(c *gin.Context) {
			unit, action := c.Param("unit"), c.Param("action")
			if !map[string]bool{"start": true, "stop": true, "restart": true, "enable": true, "disable": true}[action] {
				c.JSON(400, gin.H{"error": "invalid action"}); return
			}
			if err := collector.ServiceAction(unit, action); err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
			c.JSON(200, gin.H{"ok": true})
		})
		auth.GET("/services/:unit/logs", func(c *gin.Context) {
			lines, _ := strconv.Atoi(c.DefaultQuery("lines", "200"))
			logs, err := collector.GetServiceLogs(c.Param("unit"), lines)
			if err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
			c.JSON(200, gin.H{"logs": logs})
		})
		auth.GET("/services/:unit/file", func(c *gin.Context) {
			content, path, err := collector.ReadServiceFile(c.Param("unit"))
			if err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
			c.JSON(200, gin.H{"content": content, "path": path})
		})
		auth.POST("/services/:unit/file", func(c *gin.Context) {
			var req struct { Content string `json:"content"` }
			if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"error": "invalid request"}); return }
			if err := collector.WriteServiceFile(c.Param("unit"), req.Content); err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
			c.JSON(200, gin.H{"ok": true})
		})

		auth.GET("/metrics/history", func(c *gin.Context) {
			hours, _ := strconv.Atoi(c.DefaultQuery("hours", "24"))
			data, err := store.GetMetricsHistory(db, hours)
			if err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
			if data == nil { data = []map[string]interface{}{} }
			c.JSON(200, data)
		})

		// Settings: change username/password
		auth.POST("/settings/credentials", func(c *gin.Context) {
			var req struct {
				Username    string `json:"username"`
				Password    string `json:"password"`
				NewUsername string `json:"new_username"`
				NewPassword string `json:"new_password"`
			}
			if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"error": "invalid request"}); return }
			if req.Username != cfg.Username || req.Password != cfg.Password {
				c.JSON(401, gin.H{"error": "current credentials incorrect"}); return
			}
			if req.NewUsername != "" { cfg.Username = req.NewUsername }
			if req.NewPassword != "" { cfg.Password = req.NewPassword }
			if configPath != "" { cfg.Save(configPath) }
			c.JSON(200, gin.H{"ok": true})
		})
	}

	// Serve embedded SPA
	distFS, err := fs.Sub(webFS, "web/dist")
	if err == nil {
		r.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			if path == "/" || path == "" { path = "/index.html" }
			content, err := fs.ReadFile(distFS, strings.TrimPrefix(path, "/"))
			if err != nil {
				content, _ = fs.ReadFile(distFS, "index.html")
				c.Data(200, "text/html; charset=utf-8", content)
				return
			}
			c.Data(200, mimeType(path), content)
		})
	}
	return r
}

func loginHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"error": "invalid request"}); return }
		if req.Username != cfg.Username || req.Password != cfg.Password {
			c.JSON(401, gin.H{"error": "invalid credentials"}); return
		}
		tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": req.Username,
			"exp":      time.Now().Add(24 * time.Hour).Unix(),
		})
		tokenStr, _ := tok.SignedString([]byte(cfg.JWTSecret))
		c.JSON(200, gin.H{"token": tokenStr, "username": cfg.Username})
	}
}

func mimeType(path string) string {
	switch {
	case strings.HasSuffix(path, ".html"): return "text/html; charset=utf-8"
	case strings.HasSuffix(path, ".js"):   return "application/javascript"
	case strings.HasSuffix(path, ".css"):  return "text/css"
	case strings.HasSuffix(path, ".svg"):  return "image/svg+xml"
	case strings.HasSuffix(path, ".png"):  return "image/png"
	case strings.HasSuffix(path, ".ico"):  return "image/x-icon"
	case strings.HasSuffix(path, ".woff2"): return "font/woff2"
	}
	return "application/octet-stream"
}

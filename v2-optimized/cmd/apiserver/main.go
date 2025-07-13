package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dot/v2-optimized/internal/config"
	"dot/v2-optimized/internal/discovery"
	"dot/v2-optimized/internal/loadbalancer"
	"dot/v2-optimized/pkg/api"
)

// APIServer API服务器
type APIServer struct {
	config       *config.Config
	registry     *discovery.Registry
	loadBalancer *loadbalancer.BalancerManager
	server       *http.Server
}

// NewAPIServer 创建API服务器
func NewAPIServer(cfg *config.Config) *APIServer {
	return &APIServer{
		config:       cfg,
		registry:     discovery.NewRegistry(),
		loadBalancer: loadbalancer.NewBalancerManager(loadbalancer.Algorithm(cfg.LoadBalancer.Algorithm)),
	}
}

// Start 启动服务器
func (s *APIServer) Start() error {
	// 设置路由
	mux := http.NewServeMux()
	
	// 对象存储API
	mux.HandleFunc("/objects/", s.handleObjects)
	
	// 健康检查API
	mux.HandleFunc("/health", s.handleHealth)
	
	// 服务发现API
	mux.HandleFunc("/services", s.handleServices)
	
	// 监控API
	mux.HandleFunc("/metrics", s.handleMetrics)
	
	// 创建HTTP服务器
	s.server = &http.Server{
		Addr:         s.config.GetServiceAddress(),
		Handler:      s.loggingMiddleware(s.corsMiddleware(mux)),
		ReadTimeout:  s.config.Service.Timeout,
		WriteTimeout: s.config.Service.Timeout,
	}
	
	// 启动健康检查
	s.registry.StartHealthCheck()
	
	log.Printf("API Server starting on %s", s.config.GetServiceAddress())
	
	// 启动服务器
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}
	
	return nil
}

// Stop 停止服务器
func (s *APIServer) Stop(ctx context.Context) error {
	log.Println("Shutting down API server...")
	return s.server.Shutdown(ctx)
}

// handleObjects 处理对象存储请求
func (s *APIServer) handleObjects(w http.ResponseWriter, r *http.Request) {
	// 获取可用的数据服务器
	dataServers, err := s.registry.Discover("dataserver")
	if err != nil {
		http.Error(w, "Service discovery failed", http.StatusServiceUnavailable)
		return
	}
	
	if len(dataServers) == 0 {
		http.Error(w, "No data servers available", http.StatusServiceUnavailable)
		return
	}
	
	// 选择数据服务器
	selectedServer, err := s.loadBalancer.Select(dataServers)
	if err != nil {
		http.Error(w, "Load balancer selection failed", http.StatusServiceUnavailable)
		return
	}
	
	// 记录开始时间
	startTime := time.Now()
	s.loadBalancer.IncrementActiveConns(selectedServer.ID)
	defer s.loadBalancer.DecrementActiveConns(selectedServer.ID)
	
	// 转发请求到数据服务器
	success := s.proxyRequest(w, r, selectedServer)
	
	// 更新统计信息
	responseTime := time.Since(startTime)
	s.loadBalancer.UpdateStats(selectedServer.ID, responseTime, success)
}

// proxyRequest 代理请求到数据服务器
func (s *APIServer) proxyRequest(w http.ResponseWriter, r *http.Request, server *discovery.ServiceInfo) bool {
	// 构建目标URL
	targetURL := fmt.Sprintf("http://%s:%d%s", server.Address, server.Port, r.URL.Path)
	
	// 创建新请求
	req, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return false
	}
	
	// 复制请求头
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	
	// 发送请求
	client := &http.Client{Timeout: s.config.Service.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to proxy request", http.StatusBadGateway)
		return false
	}
	defer resp.Body.Close()
	
	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	
	// 设置状态码
	w.WriteHeader(resp.StatusCode)
	
	// 复制响应体
	if _, err := api.CopyResponse(w, resp.Body); err != nil {
		log.Printf("Failed to copy response: %v", err)
		return false
	}
	
	return resp.StatusCode < 400
}

// handleHealth 健康检查
func (s *APIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "v2-optimized",
		"services":  len(s.registry.GetAllServices()),
	}
	
	api.WriteJSON(w, health)
}

// handleServices 服务发现API
func (s *APIServer) handleServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	services := s.registry.GetAllServices()
	api.WriteJSON(w, services)
}

// handleMetrics 监控指标API
func (s *APIServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	metrics := map[string]interface{}{
		"load_balancer_stats": s.loadBalancer.GetStats(),
		"service_count":       len(s.registry.GetAllServices()),
		"timestamp":          time.Now().Unix(),
	}
	
	api.WriteJSON(w, metrics)
}

// loggingMiddleware 日志中间件
func (s *APIServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// 包装ResponseWriter以捕获状态码
		wrapped := &api.ResponseWriter{ResponseWriter: w, StatusCode: 200}
		
		next.ServeHTTP(wrapped, r)
		
		log.Printf("%s %s %d %v", r.Method, r.URL.Path, wrapped.StatusCode, time.Since(start))
	})
}

// corsMiddleware CORS中间件
func (s *APIServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	// 创建API服务器
	server := NewAPIServer(cfg)
	
	// 处理优雅关闭
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		if err := server.Stop(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()
	
	// 启动服务器
	if err := server.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

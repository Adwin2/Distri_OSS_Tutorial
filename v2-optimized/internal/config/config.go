package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config 应用配置
type Config struct {
	// 服务配置
	Service ServiceConfig `json:"service"`
	
	// 注册中心配置
	Registry RegistryConfig `json:"registry"`
	
	// 存储配置
	Storage StorageConfig `json:"storage"`
	
	// 负载均衡配置
	LoadBalancer LoadBalancerConfig `json:"load_balancer"`
	
	// 监控配置
	Monitoring MonitoringConfig `json:"monitoring"`
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	Name        string        `json:"name"`
	Host        string        `json:"host"`
	Port        int           `json:"port"`
	Environment string        `json:"environment"`
	LogLevel    string        `json:"log_level"`
	Timeout     time.Duration `json:"timeout"`
}

// RegistryConfig 注册中心配置
type RegistryConfig struct {
	Address             string        `json:"address"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	ServiceTimeout      time.Duration `json:"service_timeout"`
	RetryAttempts       int           `json:"retry_attempts"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type      string `json:"type"`      // local, s3, etc.
	RootPath  string `json:"root_path"`
	MaxSize   int64  `json:"max_size"`
	Backup    bool   `json:"backup"`
	Retention int    `json:"retention"` // days
}

// LoadBalancerConfig 负载均衡配置
type LoadBalancerConfig struct {
	Algorithm           string        `json:"algorithm"`
	HealthCheckEnabled  bool          `json:"health_check_enabled"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	MaxRetries          int           `json:"max_retries"`
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
	Enabled        bool          `json:"enabled"`
	MetricsPort    int           `json:"metrics_port"`
	CollectInterval time.Duration `json:"collect_interval"`
	AlertWebhook   string        `json:"alert_webhook"`
}

// LoadConfig 加载配置
func LoadConfig(configPath string) (*Config, error) {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		// 忽略错误，环境变量可能通过其他方式设置
	}
	
	config := &Config{}
	
	// 如果提供了配置文件路径，从文件加载
	if configPath != "" {
		if err := loadFromFile(config, configPath); err != nil {
			return nil, fmt.Errorf("failed to load config from file: %w", err)
		}
	}
	
	// 从环境变量覆盖配置
	loadFromEnv(config)
	
	// 设置默认值
	setDefaults(config)
	
	// 验证配置
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	
	return config, nil
}

// loadFromFile 从文件加载配置
func loadFromFile(config *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	
	return json.Unmarshal(data, config)
}

// loadFromEnv 从环境变量加载配置
func loadFromEnv(config *Config) {
	// 服务配置
	if val := os.Getenv("SERVICE_NAME"); val != "" {
		config.Service.Name = val
	}
	if val := os.Getenv("SERVICE_HOST"); val != "" {
		config.Service.Host = val
	}
	if val := os.Getenv("SERVICE_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			config.Service.Port = port
		}
	}
	if val := os.Getenv("ENVIRONMENT"); val != "" {
		config.Service.Environment = val
	}
	if val := os.Getenv("LOG_LEVEL"); val != "" {
		config.Service.LogLevel = val
	}
	
	// 注册中心配置
	if val := os.Getenv("REGISTRY_ADDRESS"); val != "" {
		config.Registry.Address = val
	}
	
	// 存储配置
	if val := os.Getenv("STORAGE_TYPE"); val != "" {
		config.Storage.Type = val
	}
	if val := os.Getenv("STORAGE_ROOT"); val != "" {
		config.Storage.RootPath = val
	}
	
	// 负载均衡配置
	if val := os.Getenv("LB_ALGORITHM"); val != "" {
		config.LoadBalancer.Algorithm = val
	}
	
	// 监控配置
	if val := os.Getenv("MONITORING_ENABLED"); val != "" {
		config.Monitoring.Enabled = val == "true"
	}
	if val := os.Getenv("METRICS_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			config.Monitoring.MetricsPort = port
		}
	}
}

// setDefaults 设置默认值
func setDefaults(config *Config) {
	// 服务默认值
	if config.Service.Host == "" {
		config.Service.Host = "0.0.0.0"
	}
	if config.Service.Port == 0 {
		config.Service.Port = 8080
	}
	if config.Service.Environment == "" {
		config.Service.Environment = "development"
	}
	if config.Service.LogLevel == "" {
		config.Service.LogLevel = "info"
	}
	if config.Service.Timeout == 0 {
		config.Service.Timeout = 30 * time.Second
	}
	
	// 注册中心默认值
	if config.Registry.Address == "" {
		config.Registry.Address = "localhost:8500"
	}
	if config.Registry.HealthCheckInterval == 0 {
		config.Registry.HealthCheckInterval = 30 * time.Second
	}
	if config.Registry.ServiceTimeout == 0 {
		config.Registry.ServiceTimeout = 60 * time.Second
	}
	if config.Registry.RetryAttempts == 0 {
		config.Registry.RetryAttempts = 3
	}
	
	// 存储默认值
	if config.Storage.Type == "" {
		config.Storage.Type = "local"
	}
	if config.Storage.RootPath == "" {
		config.Storage.RootPath = "/tmp/storage"
	}
	if config.Storage.MaxSize == 0 {
		config.Storage.MaxSize = 1024 * 1024 * 1024 // 1GB
	}
	if config.Storage.Retention == 0 {
		config.Storage.Retention = 30 // 30 days
	}
	
	// 负载均衡默认值
	if config.LoadBalancer.Algorithm == "" {
		config.LoadBalancer.Algorithm = "round_robin"
	}
	if config.LoadBalancer.HealthCheckInterval == 0 {
		config.LoadBalancer.HealthCheckInterval = 10 * time.Second
	}
	if config.LoadBalancer.MaxRetries == 0 {
		config.LoadBalancer.MaxRetries = 3
	}
	
	// 监控默认值
	if config.Monitoring.MetricsPort == 0 {
		config.Monitoring.MetricsPort = 9090
	}
	if config.Monitoring.CollectInterval == 0 {
		config.Monitoring.CollectInterval = 15 * time.Second
	}
}

// validateConfig 验证配置
func validateConfig(config *Config) error {
	if config.Service.Name == "" {
		return fmt.Errorf("service name is required")
	}
	
	if config.Service.Port <= 0 || config.Service.Port > 65535 {
		return fmt.Errorf("invalid service port: %d", config.Service.Port)
	}
	
	if config.Storage.RootPath == "" {
		return fmt.Errorf("storage root path is required")
	}
	
	return nil
}

// GetServiceAddress 获取服务地址
func (c *Config) GetServiceAddress() string {
	return fmt.Sprintf("%s:%d", c.Service.Host, c.Service.Port)
}

// IsProduction 是否为生产环境
func (c *Config) IsProduction() bool {
	return c.Service.Environment == "production"
}

// IsDevelopment 是否为开发环境
func (c *Config) IsDevelopment() bool {
	return c.Service.Environment == "development"
}

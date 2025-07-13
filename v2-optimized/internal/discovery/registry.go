package discovery

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// ServiceInfo 服务信息
type ServiceInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Address     string            `json:"address"`
	Port        int               `json:"port"`
	Tags        []string          `json:"tags"`
	Metadata    map[string]string `json:"metadata"`
	Health      HealthStatus      `json:"health"`
	RegisterTime time.Time        `json:"register_time"`
	LastSeen    time.Time         `json:"last_seen"`
}

// HealthStatus 健康状态
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// Registry 服务注册中心
type Registry struct {
	services map[string]*ServiceInfo
	mutex    sync.RWMutex
	
	// 配置
	healthCheckInterval time.Duration
	serviceTimeout      time.Duration
}

// NewRegistry 创建新的服务注册中心
func NewRegistry() *Registry {
	return &Registry{
		services:            make(map[string]*ServiceInfo),
		healthCheckInterval: 30 * time.Second,
		serviceTimeout:      60 * time.Second,
	}
}

// Register 注册服务
func (r *Registry) Register(service *ServiceInfo) error {
	if service.ID == "" {
		return fmt.Errorf("service ID cannot be empty")
	}
	
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	service.RegisterTime = time.Now()
	service.LastSeen = time.Now()
	service.Health = HealthStatusHealthy
	
	r.services[service.ID] = service
	
	log.Printf("Service registered: %s (%s:%d)", service.Name, service.Address, service.Port)
	return nil
}

// Deregister 注销服务
func (r *Registry) Deregister(serviceID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if service, exists := r.services[serviceID]; exists {
		delete(r.services, serviceID)
		log.Printf("Service deregistered: %s", service.Name)
		return nil
	}
	
	return fmt.Errorf("service not found: %s", serviceID)
}

// Discover 发现服务
func (r *Registry) Discover(serviceName string) ([]*ServiceInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var services []*ServiceInfo
	for _, service := range r.services {
		if service.Name == serviceName && service.Health == HealthStatusHealthy {
			services = append(services, service)
		}
	}
	
	return services, nil
}

// UpdateHealth 更新服务健康状态
func (r *Registry) UpdateHealth(serviceID string, status HealthStatus) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if service, exists := r.services[serviceID]; exists {
		service.Health = status
		service.LastSeen = time.Now()
		return nil
	}
	
	return fmt.Errorf("service not found: %s", serviceID)
}

// GetAllServices 获取所有服务
func (r *Registry) GetAllServices() map[string]*ServiceInfo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	result := make(map[string]*ServiceInfo)
	for k, v := range r.services {
		result[k] = v
	}
	
	return result
}

// StartHealthCheck 启动健康检查
func (r *Registry) StartHealthCheck() {
	ticker := time.NewTicker(r.healthCheckInterval)
	go func() {
		for range ticker.C {
			r.checkExpiredServices()
		}
	}()
}

// checkExpiredServices 检查过期服务
func (r *Registry) checkExpiredServices() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	now := time.Now()
	for id, service := range r.services {
		if now.Sub(service.LastSeen) > r.serviceTimeout {
			service.Health = HealthStatusUnhealthy
			log.Printf("Service marked as unhealthy due to timeout: %s", service.Name)
			
			// 如果服务长时间不响应，自动注销
			if now.Sub(service.LastSeen) > r.serviceTimeout*2 {
				delete(r.services, id)
				log.Printf("Service auto-deregistered due to long timeout: %s", service.Name)
			}
		}
	}
}

// ToJSON 转换为JSON格式
func (s *ServiceInfo) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

// FromJSON 从JSON格式解析
func (s *ServiceInfo) FromJSON(data []byte) error {
	return json.Unmarshal(data, s)
}

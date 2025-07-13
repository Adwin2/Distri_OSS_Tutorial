package loadbalancer

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"dot/v2-optimized/internal/discovery"
)

// LoadBalancer 负载均衡器接口
type LoadBalancer interface {
	Select(services []*discovery.ServiceInfo) (*discovery.ServiceInfo, error)
	UpdateStats(serviceID string, responseTime time.Duration, success bool)
}

// Algorithm 负载均衡算法
type Algorithm string

const (
	AlgorithmRoundRobin Algorithm = "round_robin"
	AlgorithmRandom     Algorithm = "random"
	AlgorithmWeighted   Algorithm = "weighted"
	AlgorithmLeastConn  Algorithm = "least_conn"
)

// ServiceStats 服务统计信息
type ServiceStats struct {
	TotalRequests   int64         `json:"total_requests"`
	SuccessRequests int64         `json:"success_requests"`
	FailedRequests  int64         `json:"failed_requests"`
	AvgResponseTime time.Duration `json:"avg_response_time"`
	ActiveConns     int64         `json:"active_connections"`
	LastUsed        time.Time     `json:"last_used"`
}

// BalancerManager 负载均衡管理器
type BalancerManager struct {
	algorithm Algorithm
	stats     map[string]*ServiceStats
	mutex     sync.RWMutex
	
	// Round Robin 计数器
	rrCounter uint64
}

// NewBalancerManager 创建负载均衡管理器
func NewBalancerManager(algorithm Algorithm) *BalancerManager {
	return &BalancerManager{
		algorithm: algorithm,
		stats:     make(map[string]*ServiceStats),
	}
}

// Select 选择服务实例
func (bm *BalancerManager) Select(services []*discovery.ServiceInfo) (*discovery.ServiceInfo, error) {
	if len(services) == 0 {
		return nil, fmt.Errorf("no available services")
	}
	
	switch bm.algorithm {
	case AlgorithmRoundRobin:
		return bm.selectRoundRobin(services), nil
	case AlgorithmRandom:
		return bm.selectRandom(services), nil
	case AlgorithmWeighted:
		return bm.selectWeighted(services), nil
	case AlgorithmLeastConn:
		return bm.selectLeastConn(services), nil
	default:
		return bm.selectRandom(services), nil
	}
}

// selectRoundRobin 轮询算法
func (bm *BalancerManager) selectRoundRobin(services []*discovery.ServiceInfo) *discovery.ServiceInfo {
	index := atomic.AddUint64(&bm.rrCounter, 1) % uint64(len(services))
	return services[index]
}

// selectRandom 随机算法
func (bm *BalancerManager) selectRandom(services []*discovery.ServiceInfo) *discovery.ServiceInfo {
	index := rand.Intn(len(services))
	return services[index]
}

// selectWeighted 加权算法（基于成功率）
func (bm *BalancerManager) selectWeighted(services []*discovery.ServiceInfo) *discovery.ServiceInfo {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()
	
	var bestService *discovery.ServiceInfo
	var bestScore float64 = -1
	
	for _, service := range services {
		stats := bm.getOrCreateStats(service.ID)
		
		// 计算权重分数（成功率 + 响应时间权重）
		successRate := float64(stats.SuccessRequests) / float64(stats.TotalRequests+1)
		responseTimeScore := 1.0 / (float64(stats.AvgResponseTime.Milliseconds()) + 1)
		score := successRate*0.7 + responseTimeScore*0.3
		
		if score > bestScore {
			bestScore = score
			bestService = service
		}
	}
	
	if bestService == nil {
		return services[0]
	}
	
	return bestService
}

// selectLeastConn 最少连接算法
func (bm *BalancerManager) selectLeastConn(services []*discovery.ServiceInfo) *discovery.ServiceInfo {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()
	
	var bestService *discovery.ServiceInfo
	var minConns int64 = -1
	
	for _, service := range services {
		stats := bm.getOrCreateStats(service.ID)
		
		if minConns == -1 || stats.ActiveConns < minConns {
			minConns = stats.ActiveConns
			bestService = service
		}
	}
	
	if bestService == nil {
		return services[0]
	}
	
	return bestService
}

// UpdateStats 更新服务统计信息
func (bm *BalancerManager) UpdateStats(serviceID string, responseTime time.Duration, success bool) {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()
	
	stats := bm.getOrCreateStats(serviceID)
	
	stats.TotalRequests++
	if success {
		stats.SuccessRequests++
	} else {
		stats.FailedRequests++
	}
	
	// 更新平均响应时间（简单移动平均）
	if stats.TotalRequests == 1 {
		stats.AvgResponseTime = responseTime
	} else {
		stats.AvgResponseTime = time.Duration(
			(int64(stats.AvgResponseTime)*9 + int64(responseTime)) / 10,
		)
	}
	
	stats.LastUsed = time.Now()
}

// IncrementActiveConns 增加活跃连接数
func (bm *BalancerManager) IncrementActiveConns(serviceID string) {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()
	
	stats := bm.getOrCreateStats(serviceID)
	atomic.AddInt64(&stats.ActiveConns, 1)
}

// DecrementActiveConns 减少活跃连接数
func (bm *BalancerManager) DecrementActiveConns(serviceID string) {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()
	
	stats := bm.getOrCreateStats(serviceID)
	if stats.ActiveConns > 0 {
		atomic.AddInt64(&stats.ActiveConns, -1)
	}
}

// getOrCreateStats 获取或创建服务统计信息
func (bm *BalancerManager) getOrCreateStats(serviceID string) *ServiceStats {
	if stats, exists := bm.stats[serviceID]; exists {
		return stats
	}
	
	stats := &ServiceStats{
		LastUsed: time.Now(),
	}
	bm.stats[serviceID] = stats
	return stats
}

// GetStats 获取所有服务统计信息
func (bm *BalancerManager) GetStats() map[string]*ServiceStats {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()
	
	result := make(map[string]*ServiceStats)
	for k, v := range bm.stats {
		result[k] = v
	}
	
	return result
}

// SetAlgorithm 设置负载均衡算法
func (bm *BalancerManager) SetAlgorithm(algorithm Algorithm) {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()
	
	bm.algorithm = algorithm
}

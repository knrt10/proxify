package proxy

import "log"

type LoadBalancer struct {
	Port            int
	RoundRobinCount int
	Targets         []Target
}

// NewLoadBalancer create a new LoadBalancer instance
func NewLoadBalancer(port int, targets []Target) *LoadBalancer {
	return &LoadBalancer{
		Port:            port,
		RoundRobinCount: 0,
		Targets:         targets,
	}
}

// GetNextAvailableTarget returns the address of the next available target to send a
// request to, using a simple round-robin algorithm
func (lb *LoadBalancer) GetNextAvailableTarget() Target {
	target := lb.Targets[lb.RoundRobinCount%len(lb.Targets)]
	for !target.IsAlive() {
		log.Println("Target is not alive, trying different target...")
		lb.RoundRobinCount++
		target = lb.Targets[lb.RoundRobinCount%len(lb.Targets)]
	}
	lb.RoundRobinCount++

	return target
}

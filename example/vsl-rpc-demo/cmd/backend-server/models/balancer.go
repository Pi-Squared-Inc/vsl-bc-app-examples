package models

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type LoadBalancer struct {
	Mutex     sync.Mutex
	Attesters []*AttesterEndpoint
	ctx       context.Context
	cancel    context.CancelFunc
}

type AttesterEndpoint struct {
	Url         string
	isUp        bool
	queuedTasks int
	mutex       sync.RWMutex
}

type HealthChecker interface {
	Do(req *http.Request) (*http.Response, error)
}

const MAX_QUEUED_TASKS int = 10

func NewBalancer(att_endpoints []string, checker HealthChecker) *LoadBalancer {
	lb := new(LoadBalancer)
	ctx, cancel := context.WithCancel(context.Background())
	lb.ctx = ctx
	lb.cancel = cancel
	lb.Attesters = make([]*AttesterEndpoint, len(att_endpoints))
	for idx, endpoint := range att_endpoints {
		lb.Attesters[idx] = &AttesterEndpoint{
			Url:         endpoint,
			isUp:        false,
			queuedTasks: 0,
			mutex:       sync.RWMutex{},
		}
		lb.Attesters[idx].healthCheck(lb.ctx, checker)
	}
	go lb.healthCheck(checker)
	return lb
}

func (lb *LoadBalancer) Close() {
	lb.cancel()
}

// Selects the attester with the least number of queued tasks
func (lb *LoadBalancer) GetNextAttester() (*AttesterEndpoint, error) {
	lb.Mutex.Lock()
	defer lb.Mutex.Unlock()

	// We have sufficiently few attester that this doesn't hurt:
	var min_att *AttesterEndpoint
	min_tasks := MAX_QUEUED_TASKS + 1
	for _, a := range lb.Attesters {
		a.mutex.RLock()
		if a.isUp && a.queuedTasks < min_tasks {
			min_tasks = a.queuedTasks
			min_att = a
		}
		a.mutex.RUnlock()
	}
	if min_tasks >= MAX_QUEUED_TASKS {
		return nil, fmt.Errorf("our servers are at capacity, please try again later")
	}
	if min_att == nil {
		return nil, fmt.Errorf("no attester servers available")
	}
	min_att.newTask()
	return min_att, nil
}

func (lb *LoadBalancer) healthCheck(checker HealthChecker) {
	t := time.NewTicker(time.Second * 60)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			for _, a := range lb.Attesters {
				select {
				case <-lb.ctx.Done():
					return
				default:
					a.healthCheck(lb.ctx, checker)
				}
			}
		case <-lb.ctx.Done():
			return
		}
	}
}

func (a *AttesterEndpoint) newTask() {
	a.mutex.Lock()
	a.queuedTasks += 1
	a.mutex.Unlock()
}

// This function must be called on all, and only, the attesters returned by
// GetNextAttester() after their task is finished.
func (a *AttesterEndpoint) FinishTask() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.queuedTasks -= 1
	if a.queuedTasks < 0 {
		a.queuedTasks = 0
		return fmt.Errorf("inconsistent number of queued tasks (negative)")
	}
	return nil
}
func (a *AttesterEndpoint) NumTasks() int {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.queuedTasks
}

func (a *AttesterEndpoint) healthCheck(ctx context.Context, checker HealthChecker) {
	req, err := http.NewRequestWithContext(ctx, "HEAD", a.Url+"/health_check", nil)
	if err != nil {
		return
	}
	res, err := checker.Do(req)
	a.mutex.Lock()
	defer a.mutex.Unlock()
	if err != nil {
		a.isUp = false
		log.Println(err.Error())
	} else {
		a.isUp = res.StatusCode == http.StatusOK
		res.Body.Close()
	}
	if !a.isUp {
		log.Printf("%s is down\n", a.Url)
		// Retrying of failed tasks is done by server
		a.queuedTasks = 0
	}
}

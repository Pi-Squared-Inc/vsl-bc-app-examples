package models

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

// fakeChecker implements HealthChecker for tests.
// It treats URLs containing the string "up" as healthy, others as down.
type fakeChecker struct{}

func (f fakeChecker) Do(req *http.Request) (*http.Response, error) {
	if strings.Contains(req.URL.Host, "up") {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(nil)),
		}, nil
	}
	return nil, errors.New("fake health check: endpoint down!")
}

func TestNewBalancer_AllUp(t *testing.T) {
	// Test correct health check at initialization, with two up endpoints
	testURLs := []string{"http://up1", "http://up2"}
	lb := NewBalancer(testURLs, fakeChecker{})
	defer lb.Close()

	// After init, both endpoints should be marked up
	for _, a := range lb.Attesters {
		if !a.isUp {
			t.Errorf("expected endpoint %s to be up", a.Url)
		}
	}
}

func TestGetNextAttester_AllDown(t *testing.T) {
	// Test that an error happens if all attesters are down
	lb := NewBalancer([]string{"http://down1", "http://down2", "http://down3"}, fakeChecker{})
	defer lb.Close()

	// No endpoints are up
	if _, err := lb.GetNextAttester(); err == nil {
		t.Error("expected error when no attesters are available, got nil")
	}
}

func TestGetNextAttester_LeastLoad(t *testing.T) {
	// Test correct picking of endpoint with least load
	testURLs := []string{"http://up1", "http://up2"}
	lb := NewBalancer(testURLs, fakeChecker{})
	defer lb.Close()

	att1, err := lb.GetNextAttester()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if att1.Url != "http://up1" {
		t.Errorf("expected first attester to be up1, got %s", att1.Url)
	}

	att2, err := lb.GetNextAttester()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if att2.Url != "http://up2" {
		t.Errorf("expected second attester to be up2, got %s", att2.Url)
	}
}

func TestGetNextAttester_Down(t *testing.T) {
	// Test correct picking of endpoint with least load when one endpoint becomes down
	testURLs := []string{"http://up1", "http://up2"}
	lb := NewBalancer(testURLs, fakeChecker{})
	defer lb.Close()

	// Both start with 0 tasks
	att1, err := lb.GetNextAttester()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if att1.Url != "http://up1" {
		t.Errorf("expected first attester to be up1, got %s", att1.Url)
	}

	// Now att1 has 1 task; next call would normally pick att2
	// Let's make att2 be down artificially:
	lb.Attesters[1].Url = "http://down2"
	// Force a health check:
	lb.Attesters[1].healthCheck(context.Background(), fakeChecker{})

	// Now we're expecting up1 to be picked:
	att2, err := lb.GetNextAttester()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if att2.Url != "http://up1" {
		t.Errorf("expected second attester to be up1, got %s", att2.Url)
	}
}

func TestGetNextAttester_Capacity(t *testing.T) {
	// Test that an error happens if the maximum capacity of all attesters is reached
	lb := NewBalancer([]string{"http://up"}, fakeChecker{})
	defer lb.Close()
	// Fill to capacity
	for range MAX_QUEUED_TASKS {
		if _, err := lb.GetNextAttester(); err != nil {
			t.Fatalf("unexpected error before capacity: %v", err)
		}
	}

	// Now full: next should error
	if _, err := lb.GetNextAttester(); err == nil {
		t.Error("expected capacity error, got nil")
	}
}

func TestNumTasksAndFinish(t *testing.T) {
	// Test new tasks / finished tasks to be correctly coutned
	a := &AttesterEndpoint{}

	a.newTask()
	a.newTask()

	if a.NumTasks() != 2 {
		t.Errorf("expected 2 tasks, got %d", a.NumTasks())
	}

	err := a.FinishTask()
	if err != nil {
		t.Errorf("unexpected error on FinishTask: %v", err)
	}
	if a.NumTasks() != 1 {
		t.Errorf("expected 1 task after finish, got %d", a.NumTasks())
	}

	err = a.FinishTask()
	if err != nil {
		t.Errorf("unexpected error on FinishTask: %v", err)
	}
	if a.NumTasks() != 0 {
		t.Errorf("expected 0 task after finish, got %d", a.NumTasks())
	}

	err = a.FinishTask()
	if err == nil {
		t.Error("expected error for negative finish task, got nil")
	}
	if a.NumTasks() != 0 {
		t.Errorf("expected tasks to be 0 if finish called more times than newtask, got %d", a.NumTasks())
	}
}

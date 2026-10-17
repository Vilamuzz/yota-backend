package scheduler

import (
	"sync"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
)

func TestNew(t *testing.T) {
	s := New()
	if s == nil {
		t.Fatal("New() returned nil scheduler")
	}
}

func TestAddValidSpecs(t *testing.T) {
	s := New()
	valid := []string{
		"0 0 * * *",
		"@hourly",
		"@every 1h",
	}
	for _, spec := range valid {
		if err := s.Add(spec, spec, func() {}); err != nil {
			t.Errorf("Add(%q) returned error: %v", spec, err)
		}
	}
}

func TestAddInvalidSpec(t *testing.T) {
	s := New()
	if err := s.Add("not a cron spec", "invalid", func() {}); err == nil {
		t.Error("Add() with invalid spec returned nil error")
	}
}

func TestSchedulerRunsJob(t *testing.T) {
	s := New()

	if err := s.Add("@every 100ms", "every-100ms", func() {}); err != nil {
		t.Fatalf("Add(%q) returned error: %v", "@every 100ms", err)
	}

	done := make(chan struct{})
	var mu sync.Mutex
	stable := 0
	if err := s.Add("@every 1s", "every-1s", func() {
		mu.Lock()
		stable++
		mu.Unlock()
		close(done)
	}); err != nil {
		t.Fatalf("Add(%q) returned error: %v", "@every 1s", err)
	}

	s.Start()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		s.Stop()
		t.Fatal("expected job to run within 3s")
	}
	s.Stop()

	mu.Lock()
	defer mu.Unlock()
	if stable != 1 {
		t.Errorf("expected exactly 1 execution, got %d", stable)
	}
}

func TestAddAfterStart(t *testing.T) {
	s := New()
	s.Start()
	defer s.Stop()

	if err := s.Add("@every 1m", "after-start", func() {}); err != nil {
		t.Errorf("Add() after Start returned error: %v", err)
	}
}

func TestStopWithoutStart(t *testing.T) {
	s := New()
	s.Stop()
	_ = cron.Every
}
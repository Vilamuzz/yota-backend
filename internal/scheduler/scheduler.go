package scheduler

import (
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type Scheduler struct {
	cron *cron.Cron
}

func New() *Scheduler {
	return &Scheduler{
		cron: cron.New(),
	}
}

// Add registers a new job with the given cron expression.
// Cron expression format: "sec min hour day month weekday" or standard 5-field "min hour day month weekday"
// Examples:
//
//	"0 0 * * *"   — every midnight
//	"@hourly"     — every hour
//	"@every 1h"   — every 1 hour
func (s *Scheduler) Add(spec string, name string, job func()) error {
	_, err := s.cron.AddFunc(spec, func() {
		logrus.Infof("Scheduler: running job [%s]", name)
		job()
		logrus.Infof("Scheduler: job [%s] completed", name)
	})
	if err != nil {
		logrus.Errorf("Scheduler: failed to register job [%s]: %v", name, err)
	}
	return err
}

func (s *Scheduler) Start() {
	s.cron.Start()
	logrus.Info("Scheduler: started")
}

func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	logrus.Info("Scheduler: stopped")
}

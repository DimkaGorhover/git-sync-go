package scheduler

import (
	"context"
	"github.com/procyon-projects/chrono"
	log "github.com/sirupsen/logrus"
	"io"
	"sync"
	"time"
)

type Scheduler interface {
	io.Closer
	Execute(func() error)
	Schedule(func() error, time.Duration)
	WaitError() error
}

func NewAppScheduler(errChan chan error) Scheduler {
	executor := chrono.NewDefaultTaskExecutor()
	scheduler := chrono.NewSimpleTaskScheduler(executor)
	return &appScheduler{
		scheduler: scheduler,
		errChan:   errChan,
	}
}

type appScheduler struct {
	scheduler chrono.TaskScheduler
	errChan   chan error
	wg        sync.WaitGroup
}

func (a *appScheduler) WaitError() error {
	return <-a.errChan
}

func (a *appScheduler) String() string {
	return "appScheduler"
}

func (a *appScheduler) Execute(f func() error) {
	_, err := a.scheduler.Schedule(func(ctx context.Context) {
		a.wg.Add(1)
		if err := f(); err != nil {
			a.errChan <- err
		}
		a.wg.Done()
	})
	if err != nil {
		log.WithError(err).Errorf(`error while executing task`)
	}
}

func (a *appScheduler) Schedule(f func() error, d time.Duration) {
	now := time.Now()
	startTime := now.Add(d)
	_, err := a.scheduler.ScheduleWithFixedDelay(func(ctx context.Context) {
		a.wg.Add(1)
		if err := f(); err != nil {
			a.errChan <- err
		}
		a.wg.Done()
	}, d, chrono.WithTime(startTime))

	if err != nil {
		a.errChan <- err
	}
}

func (a *appScheduler) Close() error {
	<-a.scheduler.Shutdown()
	a.wg.Wait()
	return nil
}

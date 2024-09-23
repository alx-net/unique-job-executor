package uniquejob

import (
	"context"
	"errors"
	"sync"
)

type Job[R any, K comparable] struct {
	identifier   K
	executor     ExecutorFunc[R]
	subscribers  []*Subscription[R]
	subscription *Subscription[R]
	subscribable bool
	m            sync.RWMutex
}

func NewJob[R any, K comparable](identifier K, executor ExecutorFunc[R]) *Job[R, K] {
	job := &Job[R, K]{
		identifier:   identifier,
		executor:     executor,
		subscribers:  make([]*Subscription[R], 0, 6),
		subscription: NewSubscription[R](),
		subscribable: true,
	}

	// Each job has to subscribe to itself
	job.subscribers = append(job.subscribers, job.subscription)

	return job
}

func (job *Job[R, K]) registerSubscription(subscription *Subscription[R]) error {

	job.m.Lock()
	defer job.m.Unlock()

	if !job.subscribable {
		return errors.New("can't subscribe anymore, job is already done")
	}

	job.subscribers = append(job.subscribers, subscription)

	return nil
}

func (job *Job[R, K]) run(ctx context.Context, onFinish func(s *Job[R, K])) {
	result, err := job.executor(ctx)

	onFinish(job)

	go job.sendResultToSubscribers(result, err)
}

func (job *Job[R, K]) sendResultToSubscribers(result R, err error) {

	job.m.Lock()
	defer job.m.Unlock()

	job.subscribable = false

	for _, sub := range job.subscribers {
		sub.send(result, err)
	}
}

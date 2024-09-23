package uniquejob

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Subscription[R any] struct {
	Result  chan R
	Error   chan error
	timeout *time.Timer
	m       sync.Mutex
}

func NewSubscription[R any]() *Subscription[R] {
	return &Subscription[R]{
		Result:  make(chan R),
		Error:   make(chan error),
		timeout: time.NewTimer(20 * time.Second),
	}
}

func (s *Subscription[R]) Subscribe(ctx context.Context) (R, error) {

	var res R

	defer s.close()

	select {
	case res = <-s.Result:
		return res, nil
	case err := <-s.Error:
		return res, err
	case <-ctx.Done():
		return res, errors.New("context closed")
	case <-s.timeout.C:
		return res, errors.New("request subscription timed out")
	}
}

func (s *Subscription[R]) send(result R, err error) {
	s.m.Lock()
	defer s.m.Unlock()

	select {
	case <-s.Result:
		return
	case <-s.Error:
		return
	default:
	}

	if err != nil {
		s.Error <- err
		return
	}

	s.Result <- result

}

func (s *Subscription[R]) close() {
	s.m.Lock()
	defer s.m.Unlock()

	close(s.Result)
	close(s.Error)
}

type Job[R any, K comparable] struct {
	Identifier   K
	executor     func() (R, error)
	Subscribers  []*Subscription[R]
	Subscription *Subscription[R]
	subscribable bool
	m            sync.RWMutex
}

func NewJob[R any, K comparable](identifier K, executor func() (R, error)) *Job[R, K] {
	job := &Job[R, K]{
		Identifier:   identifier,
		executor:     executor,
		Subscribers:  make([]*Subscription[R], 0, 6),
		Subscription: NewSubscription[R](),
		subscribable: true,
	}

	// Each job has to subscribe to itself
	job.Subscribers = append(job.Subscribers, job.Subscription)

	return job
}

func (job *Job[R, K]) Subscribe(ctx context.Context, subscription *Subscription[R]) error {

	job.m.Lock()
	defer job.m.Unlock()

	if !job.subscribable {
		return errors.New("can't subscribe anymore, job is already done")
	}
	job.Subscribers = append(job.Subscribers, subscription)
	fmt.Println("Subs", len(job.Subscribers))

	return nil
}

func (job *Job[R, K]) Run(ctx context.Context, onFinish func(s *Job[R, K])) {

	result, err := job.executor()

	onFinish(job)

	job.sendResultToSubscribers(ctx, result, err)
}

func (job *Job[R, K]) sendResultToSubscribers(ctx context.Context, result R, err error) {

	job.m.Lock()
	defer job.m.Unlock()

	job.subscribable = false

	for _, sub := range job.Subscribers {
		sub.send(result, err)
	}
}

type JobQueue[R any, K comparable] struct {
	workingSet map[K]*Job[R, K]
	m          sync.RWMutex
}

// K is a unique key type
func NewJobQueue[R any, K comparable]() *JobQueue[R, K] {
	return &JobQueue[R, K]{workingSet: make(map[K]*Job[R, K])}
}

func (j *JobQueue[R, K]) startNewJob(ctx context.Context, newJob *Job[R, K]) error {
	// Only allow one thread to register a job
	j.m.Lock()
	defer j.m.Unlock()

	// Check again that no one registered this job yet
	_, ok := j.workingSet[newJob.Identifier]

	if !ok {
		// Start the job and delete job after the job is done
		go newJob.Run(ctx, j.deleteJob)

		// Register
		j.workingSet[newJob.Identifier] = newJob

		return nil
	}

	return errors.New("job already registered")
}

func (j *JobQueue[R, K]) Register(ctx context.Context, newJob *Job[R, K]) {

	j.m.RLock()
	fmt.Println("Current jobs:", len(j.workingSet))
	job, ok := j.workingSet[newJob.Identifier]

	// In case job doesn't exists yet
	if !ok {
		j.m.RUnlock()

		err := j.startNewJob(ctx, newJob)

		// In case the new job was started successfully
		if err == nil {
			return
		}

		j.Register(ctx, newJob)

	} else {

		// Try to subscribe
		err := job.Subscribe(ctx, newJob.Subscription)

		j.m.RUnlock()

		// In case subscription failed try again
		if err != nil {
			j.Register(ctx, newJob)
		}
	}

}

func (j *JobQueue[R, K]) Execute(ctx context.Context, newJob *Job[R, K]) *Subscription[R] {
	// Register job
	j.Register(ctx, newJob)

	// Return jobs subscription
	return newJob.Subscription
}

func (j *JobQueue[R, K]) deleteJob(job *Job[R, K]) {
	j.m.Lock()
	defer j.m.Unlock()
	delete(j.workingSet, job.Identifier)
}

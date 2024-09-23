package uniquejob

import (
	"context"
	"errors"
	"sync"
)

type Subscription[R any] struct {
	result       chan R
	errorChannel chan error
	m            sync.Mutex
}

func NewSubscription[R any]() *Subscription[R] {
	return &Subscription[R]{
		result:       make(chan R),
		errorChannel: make(chan error),
	}
}

func (s *Subscription[R]) Subscribe(ctx context.Context) (R, error) {

	var res R

	defer s.close()

	select {
	case res = <-s.result:
		return res, nil
	case err := <-s.errorChannel:
		return res, err
	case <-ctx.Done():
		return res, errors.New("context closed")
	}
}

func (s *Subscription[R]) send(result R, err error) {
	s.m.Lock()
	defer s.m.Unlock()

	select {
	case <-s.result:
		return
	case <-s.errorChannel:
		return
	default:
	}

	if err != nil {
		s.errorChannel <- err
		return
	}

	s.result <- result

}

func (s *Subscription[R]) close() {
	s.m.Lock()
	defer s.m.Unlock()

	close(s.result)
	close(s.errorChannel)
}

type Job[R any, K comparable] struct {
	identifier   K
	executor     func() (R, error)
	subscribers  []*Subscription[R]
	subscription *Subscription[R]
	subscribable bool
	m            sync.RWMutex
}

func NewJob[R any, K comparable](identifier K, executor func() (R, error)) *Job[R, K] {
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

func (job *Job[R, K]) Subscribe(ctx context.Context, subscription *Subscription[R]) error {

	job.m.Lock()
	defer job.m.Unlock()

	if !job.subscribable {
		return errors.New("can't subscribe anymore, job is already done")
	}
	job.subscribers = append(job.subscribers, subscription)

	return nil
}

func (job *Job[R, K]) Run(ctx context.Context, onFinish func(s *Job[R, K])) {

	result, err := job.executor()

	onFinish(job)

	job.sendResultToSubscribers(result, err)
}

func (job *Job[R, K]) sendResultToSubscribers(result R, err error) {

	job.m.Lock()
	defer job.m.Unlock()

	job.subscribable = false

	for _, sub := range job.subscribers {
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
	_, ok := j.workingSet[newJob.identifier]

	if !ok {
		// Start the job and delete job after the job is done
		go newJob.Run(ctx, j.deleteJob)

		// Register
		j.workingSet[newJob.identifier] = newJob

		return nil
	}

	return errors.New("job already registered")
}

func (j *JobQueue[R, K]) Register(ctx context.Context, newJob *Job[R, K]) {

	j.m.RLock()
	job, ok := j.workingSet[newJob.identifier]

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
		err := job.Subscribe(ctx, newJob.subscription)

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
	return newJob.subscription
}

func (j *JobQueue[R, K]) deleteJob(job *Job[R, K]) {
	j.m.Lock()
	defer j.m.Unlock()
	delete(j.workingSet, job.identifier)
}

package uniquejob

import (
	"context"
	"errors"
	"sync"
)

type ExecutorFunc[R any] func(context.Context) (R, error)

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

type JobExecutor[R any, K comparable] struct {
	workingSet map[K]*Job[R, K]
	m          sync.RWMutex
}

// K is a unique key type
func NewJobExecutor[R any, K comparable]() *JobExecutor[R, K] {
	return &JobExecutor[R, K]{workingSet: make(map[K]*Job[R, K])}
}

func (j *JobExecutor[R, K]) startNewJob(ctx context.Context, newJob *Job[R, K]) error {
	// Only allow one thread to register a job
	j.m.Lock()
	defer j.m.Unlock()

	// Check again that no one registered this job yet
	_, ok := j.workingSet[newJob.identifier]

	if !ok {
		// Start the job and delete job after the job is done
		go newJob.run(ctx, j.deleteJob)

		// Register
		j.workingSet[newJob.identifier] = newJob

		return nil
	}

	return errors.New("job already registered")
}

func (j *JobExecutor[R, K]) register(ctx context.Context, newJob *Job[R, K]) {

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

		j.register(ctx, newJob)

	} else {

		// Try to subscribe
		err := job.registerSubscription(newJob.subscription)

		j.m.RUnlock()

		// In case subscription failed try again
		if err != nil {
			j.register(ctx, newJob)
		}
	}

}

func (j *JobExecutor[R, K]) Execute(ctx context.Context, newJob *Job[R, K]) *Subscription[R] {
	// Register job
	j.register(ctx, newJob)

	// Return jobs subscription
	return newJob.subscription
}

func (j *JobExecutor[R, K]) deleteJob(job *Job[R, K]) {
	j.m.Lock()
	defer j.m.Unlock()
	delete(j.workingSet, job.identifier)
}

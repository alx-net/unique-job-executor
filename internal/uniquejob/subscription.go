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

package uniquejob

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJobExecution(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	executor := func(context.Context) (int, error) {
		return 42, nil
	}

	job := NewJob("identifier", executor)

	job.run(ctx, func(s *Job[int, string]) {})

	res, err := job.subscription.Subscribe(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 42, res)
}

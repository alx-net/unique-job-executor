package http

import (
	"context"
	"testing"
	"time"

	"github.com/alx-net/concurrent-task-executor/internal/uniquejob"
	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	assert.Error(t, validate(map[string]string{"num": "-10"}))
	assert.Error(t, validate(map[string]string{"num": "asd"}))
	assert.NoError(t, validate(map[string]string{"num": "10"}))
}

func TestHandleJob(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	executor := JobExecutor{uniquejob.NewJobExecutor[Response, uint64]()}

	resp, err := handleJob(ctx, executor, 0, 10, isPrimeWrapper)
	assert.NoError(t, err)
	assert.Equal(t, false, resp.Result)

	resp, err = handleJob(ctx, executor, 0, 11, isPrimeWrapper)
	assert.NoError(t, err)
	assert.Equal(t, true, resp.Result)

	resp, err = handleJob(ctx, executor, 0, 7, fibonacciWrapper)
	assert.NoError(t, err)
	assert.Equal(t, int64(13), resp.Result)

}

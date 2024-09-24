package uniquejob

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJobExecutorExecute(t *testing.T) {
	executor := NewJobExecutor[int, string]()
	ctx := context.Background()
	jobIndentifier := "myJob"
	block := make(chan int)

	jobFunc1 := func(context.Context) (int, error) {
		<-block
		return 42, nil
	}

	jobFunc2 := func(context.Context) (int, error) {
		return 43, nil
	}

	myJob1 := NewJob(jobIndentifier, jobFunc1)
	myJob2 := NewJob(jobIndentifier, jobFunc2)

	subscription1 := executor.Execute(ctx, myJob1)
	subscription2 := executor.Execute(ctx, myJob2)

	block <- 1

	result1, err := subscription1.Subscribe(ctx)

	if err != nil {
		t.Errorf("subsciption error: %s", err)
	}

	result2, err := subscription2.Subscribe(ctx)

	if err != nil {
		t.Errorf("subsciption error: %s", err)
	}

	assert.Equal(t, result1, result2)
}

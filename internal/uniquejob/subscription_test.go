package uniquejob

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSubscriptionSend(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	sub := NewSubscription[int]()

	go sub.send(42, errors.New("test error"))

	res, err := sub.Subscribe(ctx)

	assert.Error(t, err)
	assert.Equal(t, 0, res)

	sub = NewSubscription[int]()

	go sub.send(42, nil)

	res, err = sub.Subscribe(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 42, res)
}

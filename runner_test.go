package async

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRunner_AllErrors(t *testing.T) {
	expectErr := errors.New("fail")
	runner := NewRunner(func(context.Context) error {
		return expectErr
	}).WaitErrors()
	err := runner.Run(context.Background())
	require.Equal(t, expectErr, err)
	require.Len(t, runner.AllErrors(), 1)
	require.Equal(t, expectErr, runner.AllErrors()[0])
}

func TestRunner_WaitErrors(t *testing.T) {
	expectErrOne := errors.New("fail")
	expectErrTwo := errors.New("fail")
	runner := NewRunner(func(context.Context) error {
		return expectErrOne
	}, func(context.Context) error {
		return expectErrTwo
	}).WaitErrors()
	err := runner.Run(context.Background())
	require.False(t, err != expectErrOne && err != expectErrTwo)
	require.Len(t, runner.AllErrors(), 2)
}

func TestRunner_Run(t *testing.T) {
	calledFist := false
	calledSecond := false
	runner := NewRunner(func(context.Context) error {
		calledFist = true
		return nil
	}, func(context.Context) error {
		calledSecond = true
		return nil
	})
	err := runner.Run(context.Background())
	require.Nil(t, err)
	require.True(t, calledFist)
	require.True(t, calledSecond)
}

func TestRunner_WithLimit(t *testing.T) {
	order := 1
	runner := NewRunner(func(context.Context) error {
		require.Equal(t, 1, order)
		order++
		return nil
	}, func(context.Context) error {
		require.Equal(t, 2, order)
		order++
		return nil
	}).WithLimit(1)
	err := runner.Run(context.Background())
	require.Nil(t, err)
}

func TestRunner_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	start := time.Now()
	runner := NewRunner(func(context.Context) error {
		cancel()
		time.Sleep(time.Minute)
		return nil
	})
	err := runner.Run(ctx)
	require.True(t, time.Since(start) < time.Minute)
	require.Equal(t, context.Canceled, err)
}

func TestRunner_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	start := time.Now()
	runner := NewRunner(func(context.Context) error {
		time.Sleep(time.Minute)
		return nil
	})
	err := runner.Run(ctx)
	require.True(t, time.Since(start) < time.Minute)
	require.Equal(t, context.DeadlineExceeded, err)
}

func TestRunner_Panic(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	runner := NewRunner(func(context.Context) error {
		panic(errors.New("test panic"))
		return nil
	})
	if err := runner.Run(ctx); err != nil {
		require.Contains(t, err.Error(), "async.Run: panic test panic")
	}
}

package async

import (
	"bytes"
	"context"
	"fmt"
	"runtime"
	"sync"
)

// Task ...
type Task func(context.Context) error

// Run provides a safe way to execute Task functions asynchronously, recovering if they panic
// and provides all error stack aiming to facilitate fail causes discovery
func Run(parent context.Context, tasks ...Task) error {
	ctx, cancel := context.WithCancel(parent)

	resultChannel := make(chan error, len(tasks))

	var wg sync.WaitGroup
	wg.Add(len(tasks))

	for _, task := range tasks {
		go func(fn func(context.Context) error) {
			defer wg.Done()
			defer safePanic(resultChannel)
			select {
			case <-ctx.Done():
				return // returning not to leak the goroutine
			case resultChannel <- fn(ctx):
				// Just do the job
			}
		}(task)
	}

	go func() {
		wg.Wait()
		cancel()
		close(resultChannel)
	}()

	for err := range resultChannel {
		if err != nil {
			cancel()
			return err
		}
	}

	return nil
}

// Just write the error to a error channel
// when a goroutine of a task panics
func safePanic(resultChannel chan<- error) {
	if r := recover(); r != nil {
		resultChannel <- wrapPanic(r)
	}
}

// Add miningful message to the error message
// for esiear debugging
func wrapPanic(recovered interface{}) error {
	var buf [16384]byte
	stack := buf[0:runtime.Stack(buf[:], false)]
	return fmt.Errorf("async.Run: panic %v\n%v", recovered, chopStack(stack, "panic("))
}

func chopStack(s []byte, panicText string) string {
	f := []byte(panicText)
	lfFirst := bytes.IndexByte(s, '\n')
	if lfFirst == -1 {
		return string(s)
	}
	stack := s[lfFirst:]
	panicLine := bytes.Index(stack, f)
	if panicLine == -1 {
		return string(s)
	}
	stack = stack[panicLine+1:]
	for i := 0; i < 2; i++ {
		nextLine := bytes.IndexByte(stack, '\n')
		if nextLine == -1 {
			return string(s)
		}
		stack = stack[nextLine+1:]
	}
	return string(s[:lfFirst+1]) + string(stack)
}

package async

import (
	"bytes"
	"fmt"
	"runtime"
	"sync"

	"golang.org/x/net/context"
)

// Run provides a safe way to execute fns's functions asynchronously, recovering if they panic
// and provides all error stack aiming to facilitate fail causes discovery
func Run(parent context.Context, fns ...func(ctx context.Context) error) error {
	cerr := make(chan error, 1)
	ctx, cancelFn := context.WithCancel(parent)

	go func() {
		var wg sync.WaitGroup
		wg.Add(len(fns))

		for _, fn := range fns {
			go func(fn func(ctx context.Context) error) {
				defer wg.Done()
				defer func() {
					if err := recover(); err != nil {
						cancelFn()
						var buf [16384]byte
						stack := buf[0:runtime.Stack(buf[:], false)]
						cerr <- fmt.Errorf("async.Run: panic %v\n%v",
							err, chopStack(stack, "panic("))
					}
				}()

				if err := fn(ctx); err != nil {
					cancelFn()
					cerr <- err
				}
			}(fn)
		}

		wg.Wait()
		cerr <- nil
	}()

	return <-cerr
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

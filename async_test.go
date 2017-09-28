package async_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"context"

	"github.com/StudioSol/async"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	ctx context.Context
)

func init() {
	ctx = context.Background()
}

func TestRun(t *testing.T) {
	Convey("Given two AsyncFunc functions returning non error", t, func() {
		var exec [2]bool
		f1 := func(_ context.Context) error {
			exec[0] = true
			return nil
		}

		f2 := func(_ context.Context) error {
			exec[1] = true
			return nil
		}

		Convey("It should be executed properly", func() {
			err := async.Run(context.Background(), f1, f2)
			So(err, ShouldBeNil)
			So(exec[0], ShouldBeTrue)
			So(exec[1], ShouldBeTrue)
		})
	})

	Convey("Given two AsyncFunc and one of them returning an error", t, func() {
		var errTest = errors.New("test error")
		f1 := func(_ context.Context) error {
			return errTest
		}

		f2 := func(_ context.Context) error {
			return nil
		}

		Convey("async.Run() should return that error", func() {
			err := async.Run(context.Background(), f1, f2)
			So(err, ShouldEqual, errTest)
		})
	})

	Convey("Given two AsyncFunc and one of them executing a panic call", t, func() {
		f1 := func(_ context.Context) error {
			panic(errors.New("test panic"))
		}

		f2 := func(_ context.Context) error {
			return nil
		}

		Convey("async.Run() should return that panic error", func() {
			err := async.Run(context.Background(), f1, f2)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "async.Run: panic test panic")
		})
	})

	Convey("Given two AsyncFunc and one of them executing a panic call", t, func() {
		var mu sync.Mutex
		var exec [2]bool

		f1 := func(_ context.Context) error {
			exec[0] = true
			panic(errors.New("test panic"))
		}

		f2 := func(_ context.Context) error {
			time.Sleep(5 * time.Millisecond)
			mu.Lock()
			exec[1] = true
			mu.Unlock()
			return nil
		}

		Convey("The other function should not be executed if does not need it", func() {
			_ = async.Run(context.Background(), f1, f2)
			mu.Lock()
			So(exec[1], ShouldBeFalse)
			mu.Unlock()
		})
	})

	Convey("Given an AsyncFunc executing a panic call", t, func() {
		var copyCtx context.Context
		f1 := func(ctx context.Context) error {
			copyCtx = ctx
			panic(errors.New("test panic"))
		}

		Convey("It should cancel the context", func() {
			err := async.Run(context.Background(), f1)
			So(err, ShouldNotBeNil)
			<-copyCtx.Done()
			So(copyCtx.Err(), ShouldNotBeNil)
		})
	})

	Convey("Given a cancellable context", t, func() {
		ctx, cancel := context.WithCancel(context.TODO())
		Convey("When cancelled", func() {
			cancel()
			Convey("It should cancel its children as well", func() {
				var childCtx context.Context
				f1 := func(ctx context.Context) error {
					childCtx = ctx
					return nil
				}
				err := async.Run(ctx, f1)
				So(err, ShouldBeNil)
				<-childCtx.Done()
				So(childCtx.Err(), ShouldNotBeNil)
			})
		})
	})
}

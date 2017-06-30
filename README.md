## Async

[![Build Status](https://travis-ci.org/StudioSol/async.svg?branch=master)](https://travis-ci.org/StudioSol/async)
[![codecov](https://codecov.io/gh/StudioSol/async/branch/master/graph/badge.svg)](https://codecov.io/gh/StudioSol/async)
[![Go Report Card](https://goreportcard.com/badge/github.com/StudioSol/async)](https://goreportcard.com/report/github.com/StudioSol/async)
[![GoDoc](https://godoc.org/github.com/StudioSol/async?status.svg)](https://godoc.org/github.com/StudioSol/async)

Provides a safe way to execute `fns`'s functions asynchronously, recovering them in case of panic. It also provides an error stack aiming to facilitate fail causes discovery.

### Usage
```go
func InsertAsynchronously(ctx context.Context) error {
	transaction := db.Transaction().Begin()

	err := async.Run(ctx,
		func(_ context.Context) error {
			_, err := transaction.Exec(`
				INSERT INTO foo (bar)
				VALUES ('Hello')
			`)

			return err
		},

		func(_ context.Context) error {
			_, err := transaction.Exec(`
				INSERT INTO foo (bar)
				VALUES ('world')
			`)

			return err
		},

		func(_ context.Context) error {
			_, err := transaction.Exec(`
				INSERT INTO foo (bar)
				VALUES ('asynchronously!')
			`)

			return err
		},
	)

	if err != nil {
		e := transaction.Rollback()
		log.IfError(e)
		return err
	}

	return transaction.Commit()
}

```

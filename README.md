## Async

[![Build Status](https://travis-ci.org/StudioSol/async.svg?branch=master)](https://travis-ci.org/StudioSol/async)
[![codecov](https://codecov.io/gh/StudioSol/async/branch/master/graph/badge.svg)](https://codecov.io/gh/StudioSol/async)
[![Go Report Card](https://goreportcard.com/badge/github.com/StudioSol/async)](https://goreportcard.com/report/github.com/StudioSol/async)
[![GoDoc](https://godoc.org/github.com/StudioSol/async?status.svg)](https://godoc.org/github.com/StudioSol/async)

Provides a safe way to execute functions asynchronously, recovering them in case of panic. It also provides an error stack aiming to facilitate fail causes discovery, and a simple way to control execution flow without `WaitGroup`.

### Usage
```go
var (
    user   User
    songs  []Songs
    photos []Photos
)

err := async.Run(ctx,
    func(ctx context.Context) error {
        user, err = user.Get(ctx, id)
        return err
    },
    func(ctx context.Context) error {
        songs, err = song.GetByUserID(ctx, id)
        return err
    },
    func(ctx context.Context) error {
        photos, err = photo.GetByUserID(ctx, id)
        return err
    },
)

if err != nil {
    log.Error(err)
}
```

You can also limit the number of asynchronous tasks

```go
runner := async.NewRunner(tasks...).WithLimit(3)
if err := runner.Run(ctx); err != nil { 
    log.Error(e)
}
```

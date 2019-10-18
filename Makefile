dependency:
	go get -u github.com/FiloSottile/gvt
	gvt restore

test:
	echo "" > coverage.txt
	for d in $(shell go list ./... | grep -v vendor); do \
		go test -race -v -coverprofile=profile.out -covermode=atomic $$d || exit 1; \
		[ -f profile.out ] && cat profile.out >> coverage.txt && rm profile.out; \
	done

lint:
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
	golangci-lint run --enable golint --enable gocyclo

package snitch

//go:generate /bin/bash -c "echo 'Run linters'"
//go:generate /bin/bash -c "golangci-lint run -c .golangci.yml"
//go:generate /bin/bash -c "echo 'Run go tools'"
//go:generate /bin/bash -c "goimports -e -w -format-only `find . -type f -name '*.go'`"

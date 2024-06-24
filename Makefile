test:
	go run github.com/onsi/ginkgo/v2/ginkgo -r

build:
	go build -o bin/kratix main.go

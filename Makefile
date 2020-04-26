all: test build

build: format build_native

build_with_docker:
	docker run -ti --rm -v "$$PWD:/usr/src/myapp" -v "$$HOME/go:/go" -v ~/.ssh:/root/.ssh -w /usr/src/myapp golang make build_in_docker

build_in_docker:
	bash -c 'echo -e "[url \"git@github.com:\"]\n\tinsteadOf = https://github.com/" >> /root/.gitconfig && make build_native'

build_native:
	go build -v

test: syntaxtest integrationtest coverage

syntaxtest: golint goformat

golint:
	golint . pkg/*

goformat:
	gofmt -s -w .

integrationtest:
	intergrationtest.sh

coverage:
	bash -c 'cd pkg ; go test -cover ./...'


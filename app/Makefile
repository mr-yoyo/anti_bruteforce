GO_BASE=$(shell pwd)
GO_CMD=${GO_BASE}/cmd
GO_BIN=${GO_BASE}/bin/app

build:
	CGO_ENABLED=0 go build -mod vendor -o ${GO_BIN} ${GO_CMD}

run:
	docker-compose -f ../docker-compose up -d

api_test:
	docker-compose -f ../docker-compose.test.yml up -d
	docker logs -f api_test
	docker-compose -f ../docker-compose.test.yml down

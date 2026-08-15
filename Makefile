.PHONY: frontend build test vet fmt run docker-build

frontend:
	cd frontend && npm ci && npm run build
	rm -rf internal/web/static
	mkdir -p internal/web/static
	cp -R frontend/dist/. internal/web/static/

build: frontend
	go build ./cmd/telecom

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w cmd internal

run:
	go run ./cmd/telecom

docker-build:
	docker build -t telecom:local .

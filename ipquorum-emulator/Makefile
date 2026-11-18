.PHONY: help build jar test vet integration integration-keep clean

help:
	@echo "make test         Go unit tests (emulator, -race)"
	@echo "make jar          build the fake ip_quorum.jar template (needs a JDK)"
	@echo "make build        build bin/mock-flashsystem"
	@echo "make integration  run the container integration tests (podman or docker)"
	@echo "make clean"

jar:
	./fake-ipquorum-jar/build.sh

build:
	CGO_ENABLED=0 go build -trimpath -o bin/mock-flashsystem ./cmd/mock-flashsystem

vet:
	gofmt -l . | (! grep .)
	go vet ./...

test: vet
	go test -race -count=1 ./...

integration:
	./tests/run-integration.sh

integration-keep:
	./tests/run-integration.sh --keep

clean:
	rm -rf bin fake-ipquorum-jar/build

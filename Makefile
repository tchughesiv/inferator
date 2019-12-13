# kernel-style V=1 build verbosity
ifeq ("$(origin V)", "command line")
       BUILD_VERBOSE = $(V)
endif

ifeq ($(BUILD_VERBOSE),1)
       Q =
else
       Q = @
endif

.PHONY: all
all: build

.PHONY: mod
mod:
	$(Q)go mod tidy

.PHONY: format
format: mod
	$(Q)go fmt ./...

.PHONY: go-generate
go-generate: format
	$(Q)go generate ./...

.PHONY: sdk-generate
sdk-generate: go-generate
	$(Q)operator-sdk generate k8s

.PHONY: vet
vet: sdk-generate
	$(Q)go vet ./...

.PHONY: test
test: vet
	$(Q)go test ./...

.PHONY: build
build: test
	$(Q)operator-sdk build quay.io/tchughesiv/inferator:$(shell go run getversion.go -operator)

.PHONY: csv
csv: vet
	go run ./tools/csv-gen/csv-gen.go

.PHONY: clean
clean:
	rm -rf build/_output

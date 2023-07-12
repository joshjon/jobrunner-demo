PROTO_FILES := $(wildcard proto/service/* proto/type/*)
BUF_VERSION := 1.17.0

.PHONY: run
run:
	go run cmd/main.go

.PHONY: queue
queue:
	grpcurl -plaintext -d '{"command":{"cmd": "ping", "args": ["google.com"]}}' localhost:50051 rpc.service.v1.Service/QueueJob

.PHONY: subscribe
subscribe:
	grpcurl -plaintext -d '{"job_id": "$(PARAM)"}' localhost:50051 rpc.service.v1.Service/Subscribe

stop:
	grpcurl -plaintext -d '{"job_id": "$(PARAM)"}' localhost:50051 rpc.service.v1.Service/StopJob

# Code Gen

.PHONY: gen
gen: buf-format buf-lint buf-gen

# Proto

.PHONY: buf-format
buf-format: $(PROTO_FILES)
	docker run -v $$(pwd):/srv -w /srv bufbuild/buf:$(BUF_VERSION) format -w

.PHONY: buf-lint
buf-lint: $(PROTO_FILES)
	docker run -v $$(pwd):/srv -w /srv bufbuild/buf:$(BUF_VERSION) lint

.PHONY: buf-gen
buf-gen: $(PROTO_FILES)
	rm -rf gen/
	docker run -v $$(pwd):/srv -w /srv bufbuild/buf:$(BUF_VERSION) generate

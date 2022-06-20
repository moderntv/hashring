.PHONY: test
test: TEST_RUN?=^.*$$
test: TEST_VERBOSE?=
test:
	go test \
		$(if $(TEST_VERBOSE),-v,) \
		-race \
        -timeout 1h \
		-coverprofile coverage.txt \
		-run '$(TEST_RUN)' \
		./...

.PHONY: lint
lint:
	@golangci-lint run --timeout 5m -D structcheck,unused -E bodyclose,exhaustive,exportloopref,gosec,misspell,rowserrcheck,unconvert,unparam --out-format tab --sort-results --tests=false

.PHONY: stats
stats:
	scc --exclude-dir 'vendor,node_modules,data,.git,docker/etcdkeeper,utils' --wide ./...

.PHONY: tools-install
tools-install:
	cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

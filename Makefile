GO ?= go
GOEXPERIMENT ?= jsonv2
ASSET_DIR ?= resources

.PHONY: test

test:
	env 'xray.location.asset=$(ASSET_DIR)' GOEXPERIMENT=$(GOEXPERIMENT) $(GO) test ./...

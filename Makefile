.PHONY: all reco clean fmt

GOBIN = ./build/bin
GOCMD = env GO111MODULE=on GOPROXY=https://goproxy.io go

reco:
	$(GOCMD) run build/ci.go install ./cmd/reco
	@echo "Done building."
	@echo "Run \"$(GOBIN)/reco\" to launch reco."

all:
	$(GOCMD) build -v ./...
	$(GOCMD) run build/ci.go install ./cmd/...
	@echo "Done building."
	@echo "Find binaries in \"$(GOBIN)\" directory."
	@echo ""
	@echo "Copy config-example.toml to \"$(GOBIN)\" directory"
	@cp params/config-example.toml $(GOBIN)

fmt:
	./gofmt.sh

clean:
	$(GOCMD) clean -cache
	rm -fr $(GOBIN)/*

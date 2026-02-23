GOBIN       := $(HOME)/sdk/go/bin/go
BUN         := $(HOME)/.bun/bin/bun
OUTDIR      := bin

TEAMSERVER  := $(OUTDIR)/teamserver
PUPPET      := $(OUTDIR)/puppet
DOLL        := $(OUTDIR)/doll
WHISPER     := $(OUTDIR)/whisper

.PHONY: all build build-go build-dashboard clean test lint dev-dashboard

# ── Default ───────────────────────────────────────────────────────────────────
all: build

# ── Build ─────────────────────────────────────────────────────────────────────
build: build-go build-dashboard

build-go: $(TEAMSERVER) $(PUPPET) $(DOLL) $(WHISPER)

$(OUTDIR):
	mkdir -p $(OUTDIR)

$(TEAMSERVER): $(OUTDIR) $(shell find teamserver -name '*.go')
	$(GOBIN) build -o $@ ./teamserver/cmd/teamserver/

$(PUPPET): $(OUTDIR) $(shell find puppet -name '*.go')
	$(GOBIN) build -o $@ ./puppet/cmd/puppet/

$(DOLL): $(OUTDIR) $(shell find doll -name '*.go')
	$(GOBIN) build -o $@ ./doll/cmd/doll/

$(WHISPER): $(OUTDIR) $(shell find whisper -name '*.go')
	$(GOBIN) build -o $@ ./whisper/cmd/whisper/

build-dashboard:
	cd dashboard && $(BUN) run build

# ── Test ──────────────────────────────────────────────────────────────────────
test:
	$(GOBIN) test ./...

# ── Lint ──────────────────────────────────────────────────────────────────────
lint:
	$(GOBIN) vet ./...

# ── Dev ───────────────────────────────────────────────────────────────────────
dev-dashboard:
	cd dashboard && $(BUN) run dev

# ── Clean ─────────────────────────────────────────────────────────────────────
clean:
	rm -rf $(OUTDIR)
	rm -rf dashboard/dist

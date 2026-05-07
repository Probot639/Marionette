GOBIN       := $(HOME)/sdk/go/bin/go
GARBLE      := $(HOME)/go/bin/garble
BUN         := $(HOME)/.bun/bin/bun
OUTDIR      := bin
RELDIR      := bin/release

# Workspace package list. ./... only matches the current module and the repo
# root has no go.mod, so we list each workspace member explicitly. Update when
# adding a module to go.work.
GOPKGS := ./shared/... ./teamserver/... ./puppet/... ./doll/... ./whisper/...

# Release flags: see "Anti-RE for dolls" in README.md
GARBLE_FLAGS    := -literals -tiny
RELEASE_LDFLAGS := -s -w

TEAMSERVER  := $(OUTDIR)/teamserver
PUPPET      := $(OUTDIR)/puppet
DOLL        := $(OUTDIR)/doll
WHISPER     := $(OUTDIR)/whisper

REL_TEAMSERVER := $(RELDIR)/teamserver
REL_PUPPET     := $(RELDIR)/puppet
REL_DOLL       := $(RELDIR)/doll
REL_WHISPER    := $(RELDIR)/whisper

.PHONY: all build build-go build-dashboard clean test lint dev-dashboard release release-go

all: build

# Dev build, unhardened.
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

# Hardened build. Garbled, stripped, trimmed paths. Use for anything packaged
# for an op. Requires garble: go install mvdan.cc/garble@latest
release: release-go build-dashboard

release-go: $(REL_TEAMSERVER) $(REL_PUPPET) $(REL_DOLL) $(REL_WHISPER)

$(RELDIR):
	mkdir -p $(RELDIR)

$(REL_TEAMSERVER): $(RELDIR) $(shell find teamserver -name '*.go')
	$(GARBLE) $(GARBLE_FLAGS) build -trimpath -ldflags="$(RELEASE_LDFLAGS)" -o $@ ./teamserver/cmd/teamserver/

$(REL_PUPPET): $(RELDIR) $(shell find puppet -name '*.go')
	$(GARBLE) $(GARBLE_FLAGS) build -trimpath -ldflags="$(RELEASE_LDFLAGS)" -o $@ ./puppet/cmd/puppet/

$(REL_DOLL): $(RELDIR) $(shell find doll -name '*.go')
	$(GARBLE) $(GARBLE_FLAGS) build -trimpath -ldflags="$(RELEASE_LDFLAGS)" -o $@ ./doll/cmd/doll/

$(REL_WHISPER): $(RELDIR) $(shell find whisper -name '*.go')
	$(GARBLE) $(GARBLE_FLAGS) build -trimpath -ldflags="$(RELEASE_LDFLAGS)" -o $@ ./whisper/cmd/whisper/

test:
	$(GOBIN) test $(GOPKGS)

lint:
	$(GOBIN) vet $(GOPKGS)

dev-dashboard:
	cd dashboard && $(BUN) run dev

clean:
	rm -rf $(OUTDIR)
	rm -rf dashboard/dist

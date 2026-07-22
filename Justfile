set dotenv-load := false

# ------------------------------------------------------------------
# Paths / Config
# ------------------------------------------------------------------

root            := justfile_directory()
tool_path       := root / ".tool/bin"


export PATH := tool_path + ":" + env_var("PATH")
export AQUA_ROOT_DIR := root + "/.tool"
export AQUA_DISABLE_POLICY := "true"

default:
    @just --list

shell:
    exec "${SHELL:-/bin/sh}" -i

# -----------------------------
# aqua bootstrap + installs
# -----------------------------

# dependency to ensure tool dir exists
tools-dir:
    mkdir -p "{{ tool_path }}"

# install aqua into the tools directory
aqua: tools-dir
    test -x {{tool_path}}/aqua || curl -sSfL https://raw.githubusercontent.com/aquaproj/aqua-installer/v4.0.2/aqua-installer | bash

tools *tags: (tools-aqua tags)

# install aqua based tools
tools-aqua *tags: aqua
    aqua install {{ if tags == "" { "" } else { "-t " + replace(tags, " ", " -t ") } }}

clean:
    rm -rf gen

lint: breaking
    buf lint

format:
    buf format -w

build:
    buf build

breaking:
    buf breaking --against '.git#ref=refs/remotes/origin/master' || true

update:
    #go get -u all && go mod tidy
    mkdir -p proto/openapiv3
    curl -fsSL \
      https://raw.githubusercontent.com/google/gnostic/main/openapiv3/annotations.proto \
      -o proto/openapiv3/annotations.proto

    curl -fsSL \
      https://raw.githubusercontent.com/google/gnostic/main/openapiv3/OpenAPIv3.proto \
      -o proto/openapiv3/OpenAPIv3.proto
    buf dep update

check: tools update tools clean format lint build

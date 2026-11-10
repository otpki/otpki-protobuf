set dotenv-load := false

# ------------------------------------------------------------------
# Paths / Config
# ------------------------------------------------------------------

root := justfile_directory()
tool_path := root / ".tool/bin"

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

# install aqua itself into .tool/bin — no global install needed
aqua:
    mkdir -p "{{ tool_path }}"
    test -x "{{ tool_path }}/aqua" || curl -sSfL https://raw.githubusercontent.com/aquaproj/aqua-installer/v4.1.1/aqua-installer | bash

# install every tool pinned in .aqua.yaml (or a tag subset: just tools lint)
tools *tags: aqua
    aqua install {{ if tags == "" { "" } else { "-t " + replace(tags, " ", " -t ") } }}

clean:
    rm -rf gen

lint:
    buf lint

format:
    buf format -w

alias fmt := format

# buf format diffs without rewriting; empty output means formatted.
format-check:
    buf format --diff --exit-code

build:
    buf build

breaking:
    buf breaking --against '.git#ref=refs/remotes/origin/main'

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

check: tools update clean format lint breaking build

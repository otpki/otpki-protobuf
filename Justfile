set dotenv-load := true

# ------------------------------------------------------------------
# Paths / Config
# ------------------------------------------------------------------

root            := justfile_directory()
bin_path        := root / "bin"
tool_path       := root / ".tool"

export PATH := tool_path + ":" + bin_path + ":" + env_var("PATH")

default: check

# -----------------------------
# binny bootstrap + installs
# -----------------------------

# install binny into the tools directory
binny:
    mkdir -p {{tool_path}}
    test -x {{tool_path}}/binny || curl -sSfL https://get.anchore.io/binny | sh -s -- -b {{tool_path}}

# install tools necessary to build the application
tools:
    {{tool_path}}/binny install

clean:
    rm -rf gen es

lint: breaking
    buf lint

format:
    buf format -w

build:
    buf build

breaking:
    buf breaking --against '.git#ref=refs/remotes/origin/master' || true

gen:
    buf generate
    buf generate --template buf.gen.common.yaml
    npm install && npm run generate

update:
    go get -u all && go mod tidy
    buf dep update

native-type:
    protoc \
    --go_out=. \
    --go_opt=Mproto/otpki/common/api/v1/native_type.proto=tools/protoc-gen-native/api \
    ./proto/otpki/common/api/v1/native_type.proto

check: binny update tools clean format lint build gen
    go mod tidy
    go test ./... # run a test to ensure all files compile

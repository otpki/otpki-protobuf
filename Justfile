default: check

lint:
  buf lint

format:
  buf format -w

build:
  buf build

breaking:
  buf breaking --against '.git#branch=master'

update:
  cd tools/protoc-gen-authrequired && go get -u all && go mod tidy
  cd tools/protoc-gen-native && go get -u all && go mod tidy
  buf dep update

native-type:
  protoc \
    --go_out=. \
    --go_opt=Mproto/otpki/common/api/v1/native_type.proto=tools/protoc-gen-native/api \
    ./proto/otpki/common/api/v1/native_type.proto

check: format lint build

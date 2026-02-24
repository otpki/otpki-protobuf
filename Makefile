.PHONY: lint
lint:
	buf lint

.PHONY: format
format:
	buf format -w

.PHONY: build
build:
	buf build

.PHONY: breaking
breaking:
	buf breaking --against '.git#branch=master'

.PHONY: update
update:
	cd tools/protoc-gen-authrequired && go get -u all && go mod tidy
	cd tools/protoc-gen-native && go get -u all && go mod tidy
	buf dep update

.PHONY: native-type
native-type:
	protoc \
  		--go_out=. \
  		--go_opt=Mproto/otpki/common/api/v1/native_type.proto=tools/protoc-gen-native/api \
  		./proto/otpki/common/api/v1/native_type.proto

.PHONY: check
check: format lint build
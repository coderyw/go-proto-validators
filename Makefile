# Copyright 2016 Michal Witkowski. All Rights Reserved.
# See LICENSE for licensing terms.

# 编译validator.proto
generate_validator_proto:
	@echo "--- generate validator.proto"
	protoc \
	--proto_path="$(GOPATH)/pkg/mod" --proto_path="$(GOPATH)/src" \
    --proto_path="/Users/yinwei/mygithub" \
	--gogo_out=Mgoogle/protobuf/descriptor.proto=github.com/coderyw/protobuf/protoc-gen-gogo/descriptor:. \
	validator.proto -I .

regenerate_test_gogo1: build
	@echo "--- Regenerating test .proto files with gogo imports"
	export PATH=$(extra_path):$${PATH}; protoc  \
		--proto_path="$(GOPATH)/pkg/mod" --proto_path="$(GOPATH)/src" \
        --proto_path="/Users/yinwei/mygithub" \
		--proto_path=test \
		--gogo_out=test/gogo \
		--govalidators_out=gogoimport=true:test/gogo test/*.proto

generate:
	protoc --java_out=./ --proto_path="$(GOPATH)/pkg/mod" --proto_path="$(GOPATH)/src"  --proto_path="/Users/yinwei/mygithub"  validator.proto -I .
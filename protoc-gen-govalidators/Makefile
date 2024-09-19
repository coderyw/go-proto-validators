# 测试
build:
	protoc --proto_path=../test --proto_path=../ --proto_path="$(GOPATH)/src" --proto_path="$(GOPATH)/pkg/mod" --gogo_out=../test --govalidators_out=gogoimport=true:../test ../test/*.proto
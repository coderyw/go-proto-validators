# 测试
build:
	protoc --proto_path=../test --proto_path=../ --proto_path=./ --proto_path="$(GOPATH)/pkg/mod" --gogo_out=../test/gogo --govalidators_out=gogoimport=true:../test/gogo ../test/*.proto
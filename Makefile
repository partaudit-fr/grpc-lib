test:
	go test ./...

cover:
	@GOEXPERIMENT=nocoverageredesign go test -race -coverprofile=coverage.out -covermode=atomic ./...

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.60.1 run --config scripts/.golangci.yaml

lint-fix:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.60.1 run --config scripts/.golangci.yaml --fix

clean:
ifeq ($(OS), Windows_NT)
	if exist "protogen" rd /s /q protogen
	mkdir protogen\go
else
	rm -fR ./protogen
	mkdir -p ./protogen/go
endif

protoc-go:
	cd proto && find . -name "*.proto" \
        ! -path "./google/*" \
        ! -path "./**/google/*" \
        ! -path "./protoc-gen-openapiv2/*" \
        -exec protoc -I . \
         			--go_out=../protogen/go \
                    --go_opt=paths=source_relative \
                    --go-grpc_out=../protogen/go \
                    --go-grpc_opt=paths=source_relative {} +

build: clean protoc-go

pipeline-init:
	sudo apt-get update
	sudo apt-get install -y protobuf-compiler golang-goprotobuf-dev
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

pipeline-build: pipeline-init build

clean-gateway:
ifeq ($(OS), Windows_NT)
	if exist "protogen\gateway" rd /s /q protogen\gateway
	mkdir protogen\gateway\go
else
	rm -fR ./protogen/gateway
	mkdir -p ./protogen/gateway/go
endif

clean-openapi:
ifeq ($(OS), Windows_NT)
	if exist "openapi" rd /s /q openapi
	mkdir openapi
else
	rm -fR ./protogen/gateway/openapi
	mkdir -p ./protogen/gateway/openapi
endif

protoc-go-gateway:
	cd proto && find . -name "*.proto" \
            ! -path "./google/*" \
            ! -path "./**/google/*" \
            ! -path "./protoc-gen-openapiv2/*" \
            -exec protoc -I . \
            	--grpc-gateway_out ../protogen/gateway/go \
				--grpc-gateway_opt logtostderr=true \
				--grpc-gateway_opt paths=source_relative \
				--grpc-gateway_opt standalone=true \
				--grpc-gateway_opt generate_unbound_methods=true {} +

protoc-openapiv2-gateway:
	cd proto && find . -name "*.proto" \
                ! -path "./google/*" \
                ! -path "./**/google/*" \
            	! -path "./protoc-gen-openapiv2/*" \
				-exec protoc -I . \
					--openapiv2_out ../protogen/gateway/openapi \
					--openapiv2_opt logtostderr=true \
					--openapiv2_opt output_format=yaml \
					--openapiv2_opt generate_unbound_methods=true \
					--openapiv2_opt allow_merge=true {} +

build-gateway: clean-gateway protoc-go-gateway

build-openapi: clean-openapi protoc-openapiv2-gateway

pipeline-init-gateway:
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

pipeline-build-gateway: pipeline-init-gateway build-gateway build-openapi

.PHONY: clean protoc-go build test cover lint lint-fix fix pipeline-init pipeline-build clean-gateway clean-openapi protoc-go-gateway protoc-openapiv2-gateway pipeline-init-gateway pipeline-build-gateway

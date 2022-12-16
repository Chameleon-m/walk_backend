ifneq (test, ${GIN_MODE})
    migrateCommand := migrate -source file://migrations -database "${MONGO_URI}" -verbose
else
    migrateCommand := migrate -source file://migrations -database "${MONGO_URI_TEST}" -verbose
endif

BIN_DIR = $(PWD)/bin

.PHONY: build

clean:
	rm -rf bin/*

dependencies:
	go mod download

build: dependencies build-api build-place-reindex-go-rabbitmq

build-api: 
	go build -tags ${GIN_MODE} -o ./bin/api cmd/api/main.go

build-place-reindex-go-rabbitmq:
	go build -tags ${GIN_MODE} -o ./bin/place_reindex_go_rabbitmq cmd/consumers/place_reindex_go_rabbitmq/main.go

linux-binaries:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -tags "${GIN_MODE} netgo" -installsuffix netgo -o $(BIN_DIR)/api cmd/api/main.go
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -tags "${GIN_MODE} netgo" -installsuffix netgo -o $(BIN_DIR)/place_reindex_go_rabbitmq cmd/consumers/place_reindex_go_rabbitmq/main.go

fmt: ## gofmt and goimports all go files
	find . -name '*.go' -not -wholename './vendor/*' | while read -r file; do gofmt -w -s "$$file"; goimports -w "$$file"; done

build-mocks:
	@~/go/bin/mockgen -source internal/app/repository/place_repository_interface.go -destination internal/app/repository/mock/place_repository_mock.go -package mock
	@~/go/bin/mockgen -source internal/app/repository/category_repository_interface.go -destination internal/app/repository/mock/category_repository_mock.go -package mock
	@~/go/bin/mockgen -source internal/app/repository/user_repository_interface.go -destination internal/app/repository/mock/user_repository_mock.go -package mock
	@~/go/bin/mockgen -source internal/app/service/place_service_interface.go -destination internal/app/service/mock/place_service_mock.go -package mock
	@~/go/bin/mockgen -source internal/app/service/category_service_interface.go -destination internal/app/service/mock/category_service_mock.go -package mock
	@~/go/bin/mockgen -source internal/app/service/auth_service_interface.go -destination internal/app/service/mock/auth_service_mock.go -package mock

migrate-up:
	$(migrateCommand) up $(if $n,$n,)
migrate-down:
	$(migrateCommand) down $(if $n,$n,)
migrate-goto:
	$(migrateCommand) goto $(v)
migrate-force:
	$(migrateCommand) force $(v)
migrate-drop:
	$(migrateCommand) drop
migrate-version:
	$(migrateCommand) version
migrate-create-json:
	$(migrateCommand) create -ext json -dir migrations $(name)

swagger-generate:
	swagger generate spec -o ./api/swagger.json
swagger-serve:
	swagger serve -p 8081 ./api/swagger.json
swagger-serve-f:
	swagger serve -p 8081 -F swagger ./api/swagger.json

test:
	go test -tags testing ./...

test-coverage:
	go test -tags testing ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
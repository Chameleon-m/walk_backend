ifneq (test, ${GIN_MODE})
    migrateArgs := -source file://migrations -database "${MONGO_URL}" -verbose
else
    migrateArgs := -source file://migrations -database "${MONGO_URL_TEST}" -verbose
endif

BIN_DIR = $(PWD)/bin

.PHONY: build

$(VERBOSE).SILENT:

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
	@mockgen -source internal/app/api/handlers/place/place.go -destination internal/app/api/handlers/place/mock/place.go -package mock
	@mockgen -source internal/app/api/handlers/category/category.go -destination internal/app/api/handlers/category/mock/category.go -package mock
	@mockgen -source internal/app/api/handlers/auth/auth.go -destination internal/app/api/handlers/auth/mock/auth.go -package mock
	@mockgen -source internal/app/service/place.go -destination internal/app/service/mock/place.go -package mock
	@mockgen -source internal/app/service/category.go -destination internal/app/service/mock/category.go -package mock
	@mockgen -source internal/app/service/auth.go -destination internal/app/service/mock/auth.go -package mock

migrate-up:
	migrate $(migrateArgs) up $(if $n,$n,)
migrate-down:
	migrate $(migrateArgs) down $(if $n,$n,)
migrate-goto:
	migrate $(migrateArgs) goto $(v)
migrate-force:
	migrate $(migrateArgs) force $(v)
migrate-drop:
	migrate $(migrateArgs) drop
migrate-version:
	migrate $(migrateArgs) version
migrate-create-json:
	migrate $(migrateArgs) create -ext json -dir migrations $(name)

# $(CURDIR) fix old docker version for Windows
migrate-up-docker: 
	docker run --name migrate-api --rm -i --volume="$(CURDIR)/migrations:/migrations" --network netApplication migrate/migrate:v4.15.2 $(migrateArgs) up $(if $n,$n,)
migrate-down-docker:
	docker run --name migrate-api --rm -i --volume="$(CURDIR)/migrations:/migrations" --network netApplication migrate/migrate:v4.15.2 $(migrateArgs) down $(if $n,$n,)
migrate-goto-docker:
	docker run --name migrate-api --rm -i --volume="$(CURDIR)/migrations:/migrations" --network netApplication migrate/migrate:v4.15.2 $(migrateArgs) goto $(v)
migrate-force-docker:
	docker run --name migrate-api --rm -i --volume="$(CURDIR)/migrations:/migrations" --network netApplication migrate/migrate:v4.15.2 $(migrateArgs) force $(v)
migrate-drop-docker:
	docker run --name migrate-api --rm -i --volume="$(CURDIR)/migrations:/migrations" --network netApplication migrate/migrate:v4.15.2 $(migrateArgs) drop
migrate-version-docker:
	docker run --name migrate-api --rm -i --volume="$(CURDIR)/migrations:/migrations" --network netApplication migrate/migrate:v4.15.2 $(migrateArgs) version
migrate-create-json-docker:
	docker run --name migrate-api --rm -i --volume="$(CURDIR)/migrations:/migrations" --network netApplication migrate/migrate:v4.15.2 $(migrateArgs) create -ext json -dir migrations $(name)

swagger-generate:
	swagger generate spec -o ./api/swagger.json
swagger-serve:
	swagger serve -p 8081 ./api/swagger.json
swagger-serve-f:
	swagger serve -p 8081 -F swagger ./api/swagger.json

test:
	go test -tags testing ./...
test-race:
	go test -tags -race -vet=off testing ./...
test-coverage:
	go test -tags testing ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

lints:
	golangci-lint run ./...
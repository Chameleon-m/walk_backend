ifneq (test, ${GIN_MODE})
    migrateCommand := migrate -source file://migrations -database "${MONGO_URI}" -verbose
else
    migrateCommand := migrate -source file://migrations -database "${MONGO_URI_TEST}" -verbose
endif

GOFLAGS := CGO_ENABLED=0 GOOS=linux GOARCH=amd64

.PHONY: api
api:
	${GOFLAGS} go run cmd/api/main.go

consumer-reindex-place:
	${GOFLAGS} go run cmd/consumers/place_reindex_go_rabbitmq/main.go \
	--uri="${RABBITMQ_URI}" \
	--exchange="${RABBITMQ_EXCHANGE_REINDEX}" \
	--queue="${RABBITMQ_QUEUE_PLACE_REINDEX}" \
	--binding-key="${RABBITMQ_ROUTING_PLACE_KEY}" \
	--consumer-tag="consumer_reindex_place"

generate-mocks:
	mockgen -source internal/app/repository/place_repository_interface.go -destination internal/app/repository/mock/place_repository_mock.go -package repository
	mockgen -source internal/app/repository/category_repository_interface.go -destination internal/app/repository/mock/category_repository_mock.go -package repository
	mockgen -source internal/app/service/place_service_interface.go -destination internal/app/service/mock/place_service_mock.go -package service
	mockgen -source internal/app/service/category_service_interface.go -destination internal/app/service/mock/category_service_mock.go -package service

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
ifneq (test, ${GIN_MODE})
    migrateCommand := migrate -source file://migrations -database "${MONGO_URI}" -verbose
else
    migrateCommand := migrate -source file://migrations -database "${MONGO_URI_TEST}" -verbose
endif

GOFLAGS := CGO_ENABLED=0 GOOS=linux GOARCH=amd64

api:
	${GOFLAGS} go run cmd/api/main.go

consumers:
	${GOFLAGS} go run cmd/consumers/place_reindex.go

generate-mocks:
	mockgen -source repository/place_repository_interface.go -destination repository/mock/place_repository_mock.go -package repository
	mockgen -source repository/category_repository_interface.go -destination repository/mock/category_repository_mock.go -package repository
	mockgen -source service/place_service_interface.go -destination service/mock/place_service_mock.go -package service
	mockgen -source service/category_service_interface.go -destination service/mock/category_service_mock.go -package service

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
	swagger generate spec -o ./swagger.json
swagger-serve:
	swagger serve -p 8081 ./swagger.json
swagger-serve-f:
	swagger serve -p 8081 -F swagger ./swagger.json
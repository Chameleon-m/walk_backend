FROM golang:1.20.0
WORKDIR /go/src/github.com/Chameleon-m/walk_backend
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/consumers/place_reindex_go_rabbitmq/main.go ./cmd/consumers/place_reindex_go_rabbitmq/main.go
COPY config/ ./config
COPY internal/ ./internal
WORKDIR /go/src/github.com/Chameleon-m/walk_backend/cmd/consumers/place_reindex_go_rabbitmq
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o place_reindex_go_rabbitmq .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/Chameleon-m/walk_backend/cmd/consumers/place_reindex_go_rabbitmq/place_reindex_go_rabbitmq .
CMD ["./place_reindex_go_rabbitmq"]
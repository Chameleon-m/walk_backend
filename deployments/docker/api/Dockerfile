FROM golang:1.19.4
WORKDIR /go/src/github.com/Chameleon-m/walk_backend
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/api/main.go ./cmd/api/main.go
COPY internal/ ./internal
WORKDIR /go/src/github.com/Chameleon-m/walk_backend/cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/Chameleon-m/walk_backend/cmd/api/app .
CMD ["./app"]
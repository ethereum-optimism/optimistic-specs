FROM golang:1.18.0-alpine3.15 as builder

RUN apk add --no-cache make gcc musl-dev linux-headers git

COPY Makefile /app/Makefile
COPY go.mod /app/go.mod
COPY go.sum /app/go.sum
COPY migrate.go /app/migrate.go
COPY genesis.go /app/genesis.go
WORKDIR /app
RUN go mod download
COPY ./cmd /app/cmd
RUN make surgery

FROM alpine:3.15

COPY --from=builder /app/surgery /usr/local/bin/surgery

CMD surgery

FROM golang:1.22.2-alpine3.19 as builder
WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
# RUN go build -v -o /usr/local/bin/app ./...
RUN go build -v -o /usr/local/bin/app


FROM alpine:3.19
WORKDIR /bin
COPY --from=builder /usr/local/bin/app /bin/app
CMD ["/bin/app"]
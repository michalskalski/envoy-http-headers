FROM golang:1.21 as build

ENV GO111MODULE=on
ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH amd64
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build http_server.go

FROM gcr.io/distroless/base
COPY --from=build /app/http_server /
EXPOSE 5001
ENTRYPOINT ["/http_server"]

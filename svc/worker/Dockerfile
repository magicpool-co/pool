FROM golang:1.20-buster AS build

WORKDIR /tmp/app

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod go mod download
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod CGO_ENABLED=0 go build -o worker ./svc/worker

FROM alpine:3.14 
RUN apk add ca-certificates

COPY --from=build /tmp/app/worker /app/worker

ENTRYPOINT ["/app/worker"]

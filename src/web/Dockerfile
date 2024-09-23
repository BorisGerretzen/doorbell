FROM golang:1.23.1-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Development image
FROM builder AS dev
RUN go install -mod=mod github.com/githubnemo/CompileDaemon
EXPOSE 9000
ENTRYPOINT ["CompileDaemon", "--build=go build -o /myapp ./app", "--command=/myapp", "--polling"]

# Production image
FROM builder AS build-prod
RUN go build -o /myapp ./app
FROM scratch AS prod
COPY --from=build-prod /myapp /myapp
COPY --from=build-prod /app/app/public /public
COPY --from=build-prod /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
EXPOSE 9000
ENTRYPOINT ["/myapp"]
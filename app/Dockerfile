# ---- Build stage ----
FROM golang:1.26-alpine AS builder

ARG APP_VERSION="dev"
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=${APP_VERSION}" -o app

# ---- Runtime stage: v1 current setup ----
FROM gcr.io/distroless/static-debian12:nonroot AS v1

WORKDIR /app

COPY --from=builder /app/app /app/app

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/app/app"]

# ---- Runtime stage: v2 high numeric UID/GID ----
FROM v1 AS v2

ARG HIGH_UID=100000
ARG HIGH_GID=100000

USER ${HIGH_UID}:${HIGH_GID}

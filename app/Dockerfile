# ---- Build stage ----
FROM golang:1.27-alpine@sha256:eb61c3873c97416fff0116553fff65d67a3549f72e11d9758849a689cdd8b523 AS builder

ARG APP_VERSION="dev"
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=${APP_VERSION}" -o app

# ---- Runtime stage: v1 current setup ----
FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab AS v1

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

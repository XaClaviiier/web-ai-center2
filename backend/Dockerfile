# Image backend CMS (Go) beserta web/ statis dan goose untuk migrasi. Bangun dari AKAR repo:
#   docker build -f backend/Dockerfile -t ai-center-cms:<git-sha> .
# Konfigurasi lewat environment dan berkas rahasia (*_FILE); tidak ada .env di dalam image. Lihat docs/deploy.md.
# Base image dipin ke digest agar build dapat diulang.

FROM golang:1.27.4-alpine@sha256:70bd7ddea8b6cbe5d6943cde279d6b67feb10cc924fb9100d1d8757c1575b280 AS bangun
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api \
 && CGO_ENABLED=0 GOBIN=/out go install -trimpath -ldflags="-s -w" github.com/pressly/goose/v3/cmd/goose@v3.29.0

FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab
WORKDIR /app
COPY --from=bangun /out/api /out/goose /app/
COPY backend/migrations /app/migrations
COPY web /app/web
ENV APP_ENV=production PORT=8080 STATIC_DIR=/app/web
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/api"]

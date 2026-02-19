# ============================================================
# Stage 1: Build the sypher binary
# ============================================================
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git

WORKDIR /src

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /sypher ./cmd/sypher

# ============================================================
# Stage 2: Minimal runtime image
# ============================================================
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -q --spider http://localhost:18790/health || exit 1

# Copy binary
COPY --from=builder /sypher /usr/local/bin/sypher

# Create non-root user
RUN addgroup -g 1000 sypher && \
    adduser -D -u 1000 -G sypher sypher

USER sypher

# Config dir (mounted or created at runtime)
ENV HOME=/home/sypher
RUN mkdir -p /home/sypher/.sypher-mini

EXPOSE 18790

# Entrypoint: run onboard if no config, then exec main command
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["gateway"]

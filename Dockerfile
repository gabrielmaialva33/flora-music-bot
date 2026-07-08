FROM golang:1.27rc2-bookworm AS builder

WORKDIR /build

# hadolint ignore=DL3015
RUN apt-get update && \
    apt-get install -y \
        git \
        gcc \
        unzip \
        curl \
        zlib1g-dev && \
    rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod tidy

COPY install.sh ./
COPY . .

RUN chmod +x install.sh && \
    ./install.sh -n --quiet --skip-summary && \
    CGO_ENABLED=1 go build -v -trimpath -ldflags="-w -s" -o app ./cmd/app/


FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y \
        ffmpeg \
        curl \
        unzip \
        zlib1g && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /etc/ssl/certs /etc/ssl/certs

RUN curl -fL \
      https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux \
      -o /usr/local/bin/yt-dlp && \
    chmod 0755 /usr/local/bin/yt-dlp && \
    curl -fsSL https://deno.land/install.sh -o /tmp/deno-install.sh && \
    DENO_INSTALL=/usr/local sh /tmp/deno-install.sh && \
    rm -f /tmp/deno-install.sh

# Deno resolve o desafio JS do YouTube (n-sig / signature) usado pelo yt-dlp.
# Instalado em /usr/local/bin (acessível ao usuário não-root); HOME gravável
# é necessário pro cache do solver.
ENV HOME=/home/appuser

RUN useradd -r -u 10001 -m -d /home/appuser appuser && \
    mkdir -p /app && \
    chown -R appuser:appuser /app /home/appuser

WORKDIR /app

COPY --from=builder /build/app /app/app
RUN chown appuser:appuser /app/app

USER appuser

ENTRYPOINT ["/app/app"]
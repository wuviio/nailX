# nailX Dev Container
# ベースイメージ: Microsoft 公式 Go devcontainer (Debian Bookworm)
FROM mcr.microsoft.com/devcontainers/go:1.22-bookworm

ARG DEBIAN_FRONTEND=noninteractive

# -------- システムパッケージ --------
RUN rm -f /etc/apt/sources.list.d/yarn.list && \
    apt-get update && apt-get install -y --no-install-recommends \
    curl \
    unzip \
    git \
    make \
    jq \
    ca-certificates \
    postgresql-client \
    redis-tools \
    && rm -rf /var/lib/apt/lists/*

# -------- golang-migrate --------
# DBマイグレーション実行ツール
ARG MIGRATE_VERSION=v4.17.1
RUN curl -fsSL \
    "https://github.com/golang-migrate/migrate/releases/download/${MIGRATE_VERSION}/migrate.linux-amd64.tar.gz" \
    | tar xvz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

# -------- golangci-lint --------
# Go 静的解析ツール
ARG GOLANGCI_VERSION=v1.59.1
RUN curl -sSfL \
    https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
    | sh -s -- -b /usr/local/bin ${GOLANGCI_VERSION}

# -------- Node.js 20 + Firebase CLI --------
# Firebase Emulator Suite・Firebase CLI のため
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y --no-install-recommends nodejs && \
    rm -rf /var/lib/apt/lists/* && \
    npm install -g firebase-tools@13 --no-audit --no-fund

# -------- Flutter SDK --------
# モバイル向けビルドは行わず Web・解析・テスト用途に限定。
# iOS/Android 実機デバッグはホスト側 Flutter インストールで実施する。
ARG FLUTTER_VERSION=3.22.2
RUN git clone --depth 1 --branch ${FLUTTER_VERSION} \
    https://github.com/flutter/flutter.git /opt/flutter && \
    /opt/flutter/bin/flutter config --no-analytics && \
    /opt/flutter/bin/flutter precache --web && \
    chown -R vscode:vscode /opt/flutter

ENV PATH="/opt/flutter/bin:${PATH}"
ENV FLUTTER_ROOT=/opt/flutter

# -------- Go ツール (gopls / dlv) --------
# postCreateCommand でインストールするが、ここで PATH を確認しやすくしておく
ENV GOPATH=/go
ENV PATH="${GOPATH}/bin:${PATH}"

# vscode ユーザーで動かす
USER vscode

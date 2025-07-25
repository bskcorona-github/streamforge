FROM mcr.microsoft.com/devcontainers/base:ubuntu

# 必要なパッケージのインストール
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    git \
    vim \
    nano \
    htop \
    tree \
    jq \
    yq \
    build-essential \
    pkg-config \
    libssl-dev \
    libffi-dev \
    python3-dev \
    && rm -rf /var/lib/apt/lists/*

# Go 1.22 のインストール
RUN wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz && \
    rm go1.22.0.linux-amd64.tar.gz

# Rust のインストール
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
ENV PATH="/root/.cargo/bin:${PATH}"

# Node.js 20 のインストール
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y nodejs

# Python 3.12 のインストール
RUN apt-get update && apt-get install -y \
    python3.12 \
    python3.12-venv \
    python3.12-dev \
    python3-pip \
    && rm -rf /var/lib/apt/lists/*

# kubectl のインストール
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
    chmod +x kubectl && \
    mv kubectl /usr/local/bin/

# Helm のインストール
RUN curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Docker CLI のインストール
RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg && \
    echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null && \
    apt-get update && apt-get install -y docker-ce-cli && rm -rf /var/lib/apt/lists/*

# 環境変数の設定
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:/usr/local/go/bin:$PATH
ENV CARGO_HOME=/usr/local/cargo
ENV PATH=$CARGO_HOME/bin:$PATH
ENV NODE_ENV=development

# 作業ディレクトリの設定
WORKDIR /workspace

# 非rootユーザーの作成
RUN useradd -ms /bin/bash vscode && \
    usermod -aG docker vscode && \
    chown -R vscode:vscode /workspace

# Go ツールのインストール
RUN go install golang.org/x/tools/gopls@latest && \
    go install github.com/go-delve/delve/cmd/dlv@latest && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Rust ツールのインストール
RUN rustup component add rustfmt clippy rust-analyzer

# Python ツールのインストール
RUN pip3 install --upgrade pip && \
    pip3 install black flake8 isort mypy pytest pytest-cov

# Node.js ツールのインストール
RUN npm install -g typescript @types/node ts-node nodemon

# 権限の設定
RUN chown -R vscode:vscode /go /usr/local/cargo

USER vscode 
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    curl \
    ca-certificates \
    build-essential \
    jq \
    && rm -rf /var/lib/apt/lists/*

# Install mise
RUN curl https://mise.run | sh
ENV PATH="/root/.local/bin:/root/.local/share/mise/shims:$PATH"

# Install Node.js via mise (version from .mise.toml)
COPY .mise.toml /workspace/.mise.toml
WORKDIR /workspace
RUN mise trust && mise install

# Install Claude Code (native installer; npm is deprecated)
RUN curl -fsSL https://claude.ai/install.sh | bash

# Install Codex CLI (npm downloads native Rust binary) and TAKT
RUN npm install -g @openai/codex takt

# Allow mounted volumes with different ownership
RUN git config --global --add safe.directory /workspace

CMD ["sleep", "infinity"]

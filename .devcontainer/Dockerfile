FROM golang:1.23

# Install Go tools
# Use -checklinkname=0 to disable the new linkname checks in Go 1.23
ENV GOFLAGS="-ldflags=-checklinkname=0"
RUN go install golang.org/x/tools/gopls@latest
RUN go install github.com/cweill/gotests/gotests@latest 
RUN go install github.com/fatih/gomodifytags@latest
RUN go install github.com/josharian/impl@latest
RUN go install github.com/haya14busa/goplay/cmd/goplay@latest
RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN go install honnef.co/go/tools/cmd/staticcheck@latest

# Install Node.js and PostgreSQL 16 client
RUN apt-get update && apt-get install -y \
    curl \
    gnupg \
    sudo \
    lsb-release \
    && curl -fsSL https://deb.nodesource.com/setup_18.x | bash - \
    && apt-get install -y nodejs \
    && curl -fsSL https://www.postgresql.org/media/keys/ACCC4CF8.asc | gpg --dearmor -o /usr/share/keyrings/postgresql-keyring.gpg \
    && echo "deb [signed-by=/usr/share/keyrings/postgresql-keyring.gpg] http://apt.postgresql.org/pub/repos/apt/ $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list \
    && apt-get update \
    && apt-get install -y postgresql-client-16 \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install additional development tools
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
RUN go install github.com/githubnemo/CompileDaemon@latest

# Create non-root user for GitHub Codespaces
ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID
RUN groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USER_GID -m $USERNAME \
    && echo $USERNAME ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/$USERNAME \
    && chmod 0440 /etc/sudoers.d/$USERNAME

# Set ownership of Go directories
RUN mkdir -p /go/pkg && chown -R vscode:vscode /go

# Set up workspace directory
WORKDIR /workspace
RUN chown -R vscode:vscode /workspace

# Switch to non-root user
USER vscode

# Add .bashrc configuration for better terminal experience
RUN echo "export PS1='\[\033[01;32m\]\u@\h\[\033[00m\]:\[\033[01;34m\]\w\[\033[00m\]\$ '" >> ~/.bashrc
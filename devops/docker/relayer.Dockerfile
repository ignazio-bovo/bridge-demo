# Use the official Go image
FROM golang:1.23-alpine

# Set non-interactive apt operations
ARG DEBIAN_FRONTEND=noninteractive

# Set working directory and environment variables
WORKDIR /app
ENV HOME=/app
ENV DAEMON_HOME=$HOME/.datura

# Create config directory
RUN mkdir -p $DAEMON_HOME

# Copy relayer source code
COPY ../../relayer ./

# Build the relayer
RUN go build -o relayer ./cmd/main.go && \
    cp relayer /usr/local/bin/relayer

# Set proper permissions
RUN chmod +x /usr/local/bin/relayer

# Ensure the nobody user has write permissions to the DAEMON_HOME directory
RUN chown -R nobody:nogroup $DAEMON_HOME
RUN chmod -R 0755 $DAEMON_HOME

# Switch to non-privileged user
USER nobody

# Expose the relayer ports
EXPOSE 5051
EXPOSE 5053

CMD ["relayer"]
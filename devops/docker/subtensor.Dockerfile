# Build stage
FROM rust:1.74-buster as builder

# Install dependencies
RUN apt-get update && \
    apt-get install -y \
    cmake \
    pkg-config \
    libssl-dev \
    git \
    clang \
    libclang-dev \
    protobuf-compiler

# Install rust toolchain
RUN rustup default stable && \
    rustup update && \
    rustup update nightly && \
    rustup target add wasm32-unknown-unknown --toolchain nightly

# Create working directory
WORKDIR /subtensor

# Clone the node template
RUN git clone https://github.com/opentensor/subtensor .

# Build the node and verify the binary exists
RUN cargo build --workspace --profile=release --manifest-path "./Cargo.toml"
RUN ls -la /subtensor/target/release/

# Final stage
FROM debian:buster-slim

# Install dependencies for running the node
RUN apt-get update && \
    apt-get install -y \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Copy the built binary - using the correct binary name
COPY --from=builder /subtensor/target/release/node-subtensor /usr/local/bin/

# Expose ports
EXPOSE 30333 9933 9944

# Set the entrypoint
ENTRYPOINT ["node-subtensor"]
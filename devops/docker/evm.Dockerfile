# This Dockerfile is used to build the EVM 
FROM ubuntu:20.04

SHELL ["/bin/bash", "-c"]

RUN apt update

RUN apt install -y curl git

# install foundry
RUN curl -L https://foundry.paradigm.xyz | bash

RUN /root/.foundry/bin/foundryup


# recreate the foundry directory structure inside the container
RUN mkdir /root/lib
RUN mkdir /root/script
RUN mkdir /root/src

WORKDIR /root
# Expose the default Anvil port
EXPOSE 8545

# Add environment variables with defaults
ENV ANVIL_PORT=8545
ENV ANVIL_CHAIN_ID=1

# Update CMD to use environment variables
CMD ["/bin/bash", "-c", "/root/.foundry/bin/anvil --port $ANVIL_PORT --chain-id $ANVIL_CHAIN_ID"]

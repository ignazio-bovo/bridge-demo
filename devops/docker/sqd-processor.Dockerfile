FROM node:20-bullseye-slim

# Install Python and build essentials
RUN apt-get update && apt-get install -y \
    python3 \
    make \
    g++ \
    && rm -rf /var/lib/apt/lists/*

RUN npm install -g @subsquid/cli@latest

COPY ./backend/src /app/src
COPY ./backend/db /app/db
COPY ./backend/schema.graphql /app/schema.graphql
COPY ./backend/package.json /app/package.json
COPY ./backend/tsconfig.json /app/tsconfig.json
COPY ./backend/squid.yaml /app/squid.yaml
COPY ./backend/commands.json /app/commands.json
COPY ./backend/.env /app/.env

WORKDIR /app

RUN npm install

RUN sqd codegen
RUN sqd typegen
RUN sqd build

CMD ["sqd", "run"]

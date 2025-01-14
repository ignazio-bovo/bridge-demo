## Monorepo For Datura Bridge

# Dependencies

Assumes that NodeJs v20+ and Docker are installed.

Install squid

```
npm i -g @subsquid/cli
```

# Build

1. Build backend dependency

```
cd backend
npm i
sqd build
```

2. Install dependencies on the smart contract directories

```
cd smart-contracts
npm i
```

# Run

1. Deploy smart contract only

```
npm run environment:up
```

(`npm run environment:down` for stopping)

2. Start the relayer

```
npm run relayer:start
```

(`npm run relayer:stop` for stopping)

3. Start the backend

```
npm run backend:start
```

(Ctrl+C to stop the processing and then `npm run backend:down` to bring down the db container)

## Monorepo For Datura Bridge

# Dependencies

Assumes that NodeJs v20+ and Docker are installed.

# Build

1. Install dependencies on the smart contract directories

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

(`npm run backend:stop` for stopping)

# Info For FE development

Before requesting any transfer the Token has to be whitelisted with

```
npm run contracts:whitelist_token
```

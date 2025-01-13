## Monorepo For Datura Bridge

# Dependencies

1. Install squid

```
npm i -g @subsquid/cli
```

2. Build backend dependency

```
cd backend
npm i
sqd build
```

3. Install dependencies on the smart contract directorie

```
cd smart-contracts
npm i
```

4. Run the environment

```
npm run environmen:up
```

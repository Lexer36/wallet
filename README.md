# Wallet

# config.env Configuration example (local)
- ENV=local
- SERVER_ADDRESS=:8080
- DB_DSN=postgres://postgres:postgres@localhost:5432/wallet_db?sslmode=disable
- POSTGRES_DB=wallet_db
- POSTGRES_USER=postgres
- POSTGRES_PASSWORD=postgres

# Build and run application in docker
- ```docker-compose up --build -d```

# APIs:
# 1.  POST /api/v1/wallet

Deposit or withdraw funds from a wallet.

- Request body:

```
{
    "walletId": "c8b43e22-3cc0-4647-b18b-53fba78d6fed",
    "operationType": "DEPOSIT" or "WITHDRAW",
    "amount": 1000
}
```

walletId — UUID wallet.

operationType — operation type: DEPOSIT or WITHDRAW.

amount — amount of money.

- Response: 
 ``200 OK``
- 
# 2. Get balance for a wallet
   GET /api/v1/wallets/{walletId}

- Request example

curl -X GET http://localhost:8080/api/v1/wallets/c8b43e22-3cc0-4647-b18b-53fba78d6fed

- Response

```
{
    "walletId": "c8b43e22-3cc0-4647-b18b-53fba78d6fed",
    "balance": 1000
}
```

# Migrations using Goose
-` For now migrations apply on app start from ./migrations directory`

- Manual migration UP example:
- ```goose -dir ./migrations postgres "postgres://postgres:postgres@localhost:5432/wallet_db?sslmode=disable" up```
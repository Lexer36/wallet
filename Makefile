PORT = 8080
URL = http://localhost:$(PORT)

build:
	docker-compose up --build -d

test:
	go test ./internal/...

deposit:
	curl -X POST $(URL)/api/v1/wallet \
	-H "Content-Type: application/json" \
	-d '{"walletId": {"example-wallet-id"}, "operationType": "DEPOSIT", "amount": 1000}'

withdraw:
	curl -X POST $(URL)/api/v1/wallet \
	-H "Content-Type: application/json" \
	-d '{"walletId": {"example-wallet-id"}, "operationType": "WITHDRAW", "amount": 500}'

get-balance:
	curl $(URL)/api/v1/wallets/{example-wallet-id}
# ip_checking_bot

Golang bot for ip validation

1. For starting database locally use:
```
docker run -e POSTGRES_PASSWORD=root -p 3306:5432 -d postgres
```

2. Setup the required tables using the command:
```
go run cmd/setup/setup.go
```

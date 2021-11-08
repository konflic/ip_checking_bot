# ip_checking_bot

Golang bot for ip check via external IP allows to obtain basic information about given IP adress. It also includes admin functionallity.

0. For working properly app require environment variable with API telegram token to be set as TELEGRAM_TOKEN.

1. In order to create a default admin for bot DEFAULT_ADMIN environment variable must be set with username of admin.

# Run locally without docker

For starting database locally you can setup it as docker container. App will be working with the following configuration by default.
```
docker run -e POSTGRES_PASSWORD=root -p 5432:5432 -d postgres
```
After setting variables from 0. and 1. run
```
go run cmd/bot/main.go
```

Database constant is located within helpers/dbhelpers.go file.

# Setup with docker-compose

In order to setup API, BOT with Database all at once use the following command:

```
TELEGTAM_TOKEN={your token} DEFAULT_ADMIN={your_admin} docker-compose up -d
```

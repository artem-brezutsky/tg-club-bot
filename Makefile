.PHONY:
.SILENT:

build:
	go build -o ./.bin/bot cmd/bot/main.go

run: build
	./.bin/bot

build-image:
	docker-compose build

run-dc: build-image
	docker-compose up -d

stop-dc:
	docker-compose down

run-dc-db:
	docker-compose up db -d
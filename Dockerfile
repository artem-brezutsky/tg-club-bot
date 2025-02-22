# Первый этам сборки
FROM golang:1.19-alpine3.17 as builer

COPY . /telegram_bot/
WORKDIR /telegram_bot/

RUN go mod download
RUN go build -o ./.bin/bot ./cmd/bot/main.go
RUN ls -la

FROM alpine:latest

WORKDIR /root/

#COPY --from=builer /telegram_bot/.env .
COPY --from=builer /telegram_bot/.bin/bot .
COPY --from=builer /telegram_bot/configs configs/

EXPOSE 80

CMD ["./bot"]
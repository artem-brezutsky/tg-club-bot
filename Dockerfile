# Первый этам сборки
FROM golang:1.19-alpine3.17 as builer

COPY . /bmwBot/
WORKDIR /bmwBot/

RUN go mod download
RUN go build -o ./.bin/bot ./cmd/bot/main.go
RUN ls -la

FROM alpine:latest

WORKDIR /root/

COPY --from=builer /bmwBot/.env .
COPY --from=builer /bmwBot/.bin/bot .
COPY --from=builer /bmwBot/configs configs/

EXPOSE 80

CMD ["./bot"]
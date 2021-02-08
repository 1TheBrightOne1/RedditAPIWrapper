FROM golang:1.14

COPY . ./src/github.com/1TheBrightOne1/RedditAPIWrapper

COPY ignoredStocks.txt /var/stonks/ignoredStocks.txt

COPY tickers.txt /var/stonks/tickers.txt

WORKDIR ./src/github.com/1TheBrightOne1/RedditAPIWrapper

RUN go get ./...

RUN go build main.go

ENTRYPOINT ["./main"]
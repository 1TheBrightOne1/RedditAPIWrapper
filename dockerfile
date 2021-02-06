FROM golang:1.14

COPY . ./src/github.com/1TheBrightOne1/RedditAPIWrapper

WORKDIR ./src/github.com/1TheBrightOne1/RedditAPIWrapper

RUN go get ./...

RUN go build main.go

ENTRYPOINT ["./main"]
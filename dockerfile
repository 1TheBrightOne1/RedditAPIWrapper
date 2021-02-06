FROM golang:1.14

COPY . ./src/github.com/1TheBrightOne1/RedditAPIWrapper

RUN cd ./src/github.com/1TheBrightOne1/RedditAPIWrapper ; go get ./...

RUN cd ./src/github.com/1TheBrightOne1/RedditAPIWrapper ; go build main.go

ENTRYPOINT ["cd ./src/github.com/1TheBrightOne1/RedditAPIWrapper ; ./main"]
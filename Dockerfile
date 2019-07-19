FROM golang:latest

ENV GO111MODULE off

WORKDIR /go/src/app
COPY . .

RUN go get -d ./...

RUN go install 

CMD ["app"]
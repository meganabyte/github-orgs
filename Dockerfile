FROM golang:latest

# environment vary
WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go install 

CMD ["app"]
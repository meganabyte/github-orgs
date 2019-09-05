FROM golang:latest

LABEL maintainer="Megana Bobba <megbobba@gmail.com>"

ARG AWS_ACCESS_KEY_ID
ARG AWS_SECRET_ACCESS_KEY

ENV GOPATH /go
ENV GO111MODULE off
ENV AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID:-production}
ENV AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY:-production}

WORKDIR /go/src/app
COPY ./src/github.com/meganabyte/github-orgs /go/src/app

RUN go get -d ./...
RUN go install ./...

EXPOSE 8080
CMD ["app"]

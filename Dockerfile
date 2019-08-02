FROM golang:latest

LABEL maintainer="Megana Bobba <megbobba@gmail.com>"

ARG ORG_NAME
ARG OAUTH_TOKEN
ARG USER_NAME

ENV GOPATH /go
ENV GO111MODULE off
ENV ORG_NAME=${ORG_NAME:-production}
ENV OAUTH_TOKEN=${OAUTH_TOKEN:-production}
ENV USER_NAME=${USER_NAME:-production}

WORKDIR /go/src/app
COPY ./src/github.com/meganabyte/github-orgs /go/src/app

RUN go get -d -v ./...
RUN go install -v ./...

EXPOSE 8080
CMD ["app"]

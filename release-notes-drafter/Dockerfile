# Builder container
FROM golang:1.11.4 as builder

WORKDIR /go/src/github.com/gruntwork-io/prototypes/release-notes-drafter/

# Install dep
RUN curl -O https://raw.githubusercontent.com/golang/dep/master/install.sh \
    && /bin/sh install.sh

# Build app
COPY *.go ./
COPY Gopkg.* ./
RUN dep ensure \
    && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o release-notes-drafter .


# Runtime container
FROM debian:stretch

LABEL "com.github.actions.name"="Release Notes Drafter"
LABEL "com.github.actions.description"="Draft release notes in Gruntwork style on pull request merge"
LABEL "com.github.actions.icon"="edit-3"
LABEL "com.github.actions.color"="purple"

LABEL "repository"="https://github.com/gruntwork-io/prototypes"
LABEL "homepage"="https://gruntwork.io"
LABEL "maintainer"="Gruntwork <www.gruntwork.io>"

RUN apt-get update && apt-get -y upgrade && apt-get -y install ca-certificates
COPY --from=builder /go/src/github.com/gruntwork-io/prototypes/release-notes-drafter/release-notes-drafter /release-notes-drafter
ENTRYPOINT ["/release-notes-drafter"]

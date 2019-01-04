FROM debian:stretch

LABEL "com.github.actions.name"="Release Notes Drafter"
LABEL "com.github.actions.description"="Draft release notes in Gruntwork style on pull request merge"
LABEL "com.github.actions.icon"="edit-3"
LABEL "com.github.actions.color"="purple"

LABEL "repository"="https://github.com/gruntwork-io/prototypes"
LABEL "homepage"="https://gruntwork.io"
LABEL "maintainer"="Gruntwork <www.gruntwork.io>"

RUN apt-get update && apt-get -y upgrade && apt-get -y install ca-certificates
COPY ./release-notes-drafter /release-notes-drafter
ENTRYPOINT ["/release-notes-drafter"]

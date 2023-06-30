FROM golang:1.20-alpine as build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY templates/ ./templates/

RUN apk add --no-cache build-base

#RUN CGO_ENABLED=1 go build -ldflags="-s -w -linkmode external -extldflags '-static'"
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w -linkmode external -extldflags '-static'" -o /atomstr

FROM alpine:latest

LABEL org.opencontainers.image.title="atomstr"
LABEL org.opencontainers.image.source=https://sr.ht/~psic4t/atromstr
LABEL org.opencontainers.image.description="Atomstr scrapes RSS or Atom feeds and posts them into Nostr"
LABEL org.opencontainers.image.authors="Raúl Piracés"
LABEL org.opencontainers.image.licenses=GPL

ENV FETCH_INTERVAL="15m"
ENV METADATA_INTERVAL="2h"
ENV LOG_LEVEL="INFO"
ENV WEBSERVER_PORT="8061"
ENV DB_DIR="/db/rsslay.sqlite"
ENV DEFAULT_PROFILE_PICTURE_URL="https://i.imgur.com/MaceU96.png"

COPY --from=build /atomstr .

CMD [ "/atomstr" ]

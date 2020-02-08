# build
FROM golang:alpine as builder

RUN apk update \
  && apk add git

RUN mkdir /app
WORKDIR /app
COPY go.mod .
COPY go.sum .

RUN go mod download
COPY . .

RUN go build -o /main


# final image
FROM alpine:3.10

RUN adduser -D -u 1000 server
COPY --from=builder /main /home/server/server
RUN chown -R server /home/server && \
    chmod o+x /home/server/server

USER 1000
ENTRYPOINT ["/home/server/server"]

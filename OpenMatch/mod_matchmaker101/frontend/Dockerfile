FROM golang:alpine as go
WORKDIR /app
ENV GO111MODULE=on

COPY . .
RUN go build -o frontend .

CMD ["/app/frontend"]

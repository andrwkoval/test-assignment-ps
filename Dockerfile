FROM golang:1.18.3 AS build

WORKDIR /app

ENV CGO_ENABLED=0
ENV GO111MODULE=on

COPY . .

RUN go mod download

RUN go build -o /ps-assignment-app

# Production
FROM alpine:latest

#ENV REQUESTS_PER_MINUTE_LIMIT=10

WORKDIR /

COPY --from=build /ps-assignment-app /ps-assignment-app

EXPOSE 10000

ENTRYPOINT ["/ps-assignment-app"]
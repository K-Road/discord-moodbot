FROM golang:1.24 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN go build -o moodbot

FROM ubuntu:20.04
RUN apt-get update && apt-get install -y libc6
COPY --from=builder /app/moodbot /moodbot
CMD ["/moodbot"]

FROM golang:1.24 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN go build -o moodbot

FROM gcr.io/distroless/static
COPY --from=builder /app/moodbot /
CMD ["/moodbot"]

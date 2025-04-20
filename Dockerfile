FROM golang:1.24 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o moodbot

FROM gcr.io/distroless/static
COPY --from=builder /app/moodbot /moodbot
ENV PORT=8080
CMD ["/moodbot"]

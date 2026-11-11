# syntax=docker/dockerfile:1

# Tahap 1: Build (Compile kode Go)
FROM golang:1.26-alpine AS build
WORKDIR /src

# Copy file dependency dulu agar caching lebih cepat
COPY go.mod go.sum ./
RUN go mod download

# Copy seluruh kode dan compile menjadi binary bernama 'app'
COPY . .
RUN CGO_ENABLED=0 go build -o /out/app .

# Tahap 2: Runtime (Hanya membawa binary, tanpa compiler Go)
FROM alpine:3.20

# Bikin user non-root agar lebih aman
RUN addgroup -S app && adduser -S -G app app

WORKDIR /app
# Copy binary dari tahap build
COPY --from=build /out/app ./app

# Bikin folder logs agar aplikasi bisa menulis log
RUN mkdir logs && chown app:app logs
USER app

EXPOSE 8080
ENTRYPOINT ["./app"]
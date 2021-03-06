FROM golang:alpine AS builder

WORKDIR /app

# Copy source code.
COPY *.go ./
COPY go.* ./

# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o mgott

############################

FROM alpine AS runner

# Copy our static executable and static files.
COPY --from=builder /app/mgott /app/mgott

# Move to /app because relative path is use in the binary.
WORKDIR /app

# Run the binary.
ENTRYPOINT ["/app/mgott"]

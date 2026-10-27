# Stage 1: Build WebAssembly client
FROM golang:1.24-alpine AS wasm-builder
WORKDIR /app
COPY go.mod ./
COPY . .
RUN GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o web/static/app.wasm ./web/wasm \
    && cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/static/wasm_exec.js

# Stage 2: Build Native Go Server with embedded WASM & web assets
FROM golang:1.24-alpine AS server-builder
WORKDIR /app
COPY --from=wasm-builder /app ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/gorithm ./cmd/gorithm

# Stage 3: Minimal Runtime
FROM alpine:3.21
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=server-builder /bin/gorithm /app/gorithm

EXPOSE 8080
ENV PORT=8080
CMD ["/app/gorithm"]

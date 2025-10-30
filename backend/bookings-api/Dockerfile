# ==========================================
# STAGE 1: Build
# ==========================================
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copiar go.mod y go.sum
COPY go.mod go.sum ./

# Descargar dependencias
RUN go mod download

# Copiar el código fuente
COPY . .

# Compilar la aplicación
RUN CGO_ENABLED=0 GOOS=linux go build -o bookings-api .

# ==========================================
# STAGE 2: Runtime
# ==========================================
FROM alpine:latest

WORKDIR /root/

# Instalar ca-certificates para HTTPS
RUN apk --no-cache add ca-certificates

# Copiar el binario desde el builder
COPY --from=builder /app/bookings-api .

# Exponer el puerto
EXPOSE 8082

# Ejecutar la aplicación
CMD ["./bookings-api"]
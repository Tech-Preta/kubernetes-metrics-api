# Dockerfile

# Estágio de build: usa uma imagem Go para compilar a aplicação
FROM golang:1.24-alpine AS builder

# Define o diretório de trabalho dentro do container
WORKDIR /app

# Instala certificados para conexões HTTPS
RUN apk --no-cache add ca-certificates

# Copia os arquivos de módulo e baixa as dependências primeiro
# Isso aproveita o cache do Docker se as dependências não mudarem
COPY go.mod go.sum ./
RUN go mod download

# Copia o restante do código da aplicação
COPY . .

# Compila a aplicação.
# -o /app/k8s-metrics-api: especifica o nome e local do binário de saída
# CGO_ENABLED=0: desabilita CGO para criar um binário estático (bom para Alpine)
# GOOS=linux GOARCH=amd64: especifica o sistema operacional e arquitetura alvo (comum para containers)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/k8s-metrics-api .

# Estágio final: usa uma imagem base leve (Alpine) para a imagem final
FROM alpine:3.19

# Define o diretório de trabalho
WORKDIR /app

# Cria usuário não-root (usando os comandos básicos do alpine)
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Copia apenas o binário compilado do estágio de build
COPY --from=builder /app/k8s-metrics-api ./k8s-metrics-api

# Garante que o binário seja executável e muda ownership para o usuário não-root
RUN chmod +x /app/k8s-metrics-api && chown appuser:appgroup /app/k8s-metrics-api

# Muda para o usuário não-root
USER appuser

# Expõe a porta que a aplicação usa (a mesma definida na API e no Helm chart)
EXPOSE 8080

# Define a variável de ambiente PORT que a API usa (se não for passada pelo Kubernetes)
# No entanto, é melhor configurar isso via ConfigMap/Secret no Kubernetes ou
# deixar que o Deployment passe como env var, como fizemos no Helm Chart.
# ENV PORT=8080

# Adiciona HEALTHCHECK para verificar se a aplicação está funcionando
# Usando curl interno do Alpine (se disponível) ou alternativa
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD nc -z localhost 8080 || exit 1

# Comando para executar a aplicação quando o container iniciar
# O token EXPECTED_AUTH_TOKEN será injetado pelo Kubernetes (via Secret, como no Helm chart)
CMD ["./k8s-metrics-api"]
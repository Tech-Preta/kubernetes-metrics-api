# Kubernetes API Metrics

API de métricas Kubernetes que fornece informações sobre o cluster através de endpoints RESTful protegidos por autenticação.

## Índice

- [k8s-api-metrics](#k8s-api-metrics)
  - [Índice](#índice)
  - [Visão Geral](#visão-geral)
  - [Requisitos](#requisitos)
    - [Dependências](#dependências)
  - [Histórico de Versões](#histórico-de-versões)
    - [v1.0.1 (27 de maio de 2025)](#v101-27-de-maio-de-2025)
      - [Detalhes da correção na v1.0.1](#detalhes-da-correção-na-v101)
  - [Desenvolvimento Local](#desenvolvimento-local)
    - [Compilar e Executar](#compilar-e-executar)
    - [Testar Localmente](#testar-localmente)
  - [Docker](#docker)
    - [Build da Imagem](#build-da-imagem)
    - [Executar com Docker](#executar-com-docker)
    - [Publicar Imagem no Registry](#publicar-imagem-no-registry)
  - [Kubernetes](#kubernetes)
    - [Deploy com Helm](#deploy-com-helm)
    - [Acessar a API no Kubernetes](#acessar-a-api-no-kubernetes)
    - [Atualizar o Deployment](#atualizar-o-deployment)
    - [Troubleshooting](#troubleshooting)
  - [Endpoints da API](#endpoints-da-api)
    - [Exemplos de Resposta](#exemplos-de-resposta)
      - [`/metrics` (JSON)](#metrics-json)
      - [`/healthz` (Health Check)](#healthz-health-check)
  - [Autenticação](#autenticação)
  - [Observações e Melhorias](#observações-e-melhorias)
  - [Contribuindo](#contribuindo)
    - [Validação e Testes](#validação-e-testes)
      - [1. Teste Local com Go](#1-teste-local-com-go)
      - [2. Teste com Docker](#2-teste-com-docker)
      - [3. Teste com Kubernetes](#3-teste-com-kubernetes)
    - [Convenções de Código](#convenções-de-código)
  - [Licença](#licença)

## Visão Geral

Esta API é desenvolvida em Go e coleta métricas do cluster Kubernetes onde está sendo executada. Ela fornece informações como:

- Contagem de nós, pods, deployments e serviços
- Informações detalhadas sobre cada nó
- Status dos componentes do cluster

Os dados são disponibilizados em formato JSON e também em formato Prometheus para monitoramento.

## Requisitos

- Go 1.24 ou superior (para desenvolvimento local)
- Docker (para build da imagem)
- Kubernetes (para deploy)
- Helm 3.x (para deploy usando o chart)
- Acesso a um cluster Kubernetes (para teste e implantação)

### Dependências

O projeto utiliza as seguintes dependências principais:

```go
// Arquivo go.mod
require (
	github.com/prometheus/client_golang v1.22.0
	k8s.io/apimachinery v0.33.1
	k8s.io/client-go v0.33.1
)
```

Para instalar todas as dependências, execute `go mod download` no diretório do projeto.

## Histórico de Versões

- **v1.0.0** - Lançamento inicial da API com métricas básicas do cluster.
- **v1.0.1** - Melhoria na validação do token de autenticação e atualização na documentação

### v1.0.1 (27 de maio de 2025)
- Corrigido problema de autenticação que rejeitava tokens válidos
- Adicionado `strings.TrimSpace()` para remover quebras de linha e espaços indesejados no token
- Configurado uso de variável de ambiente `EXPECTED_AUTH_TOKEN` para maior flexibilidade
- Melhorada a documentação e exemplos de uso

#### Detalhes da correção na v1.0.1

O principal problema resolvido na versão v1.0.1 foi a rejeição incorreta de tokens de autenticação válidos. A causa-raiz era:

1. O token armazenado no Secret Kubernetes tinha uma quebra de linha extra (`\n`) no final
2. A API não removia espaços em branco ou quebras de linha do token lido da variável de ambiente

A solução implementada foi:
```go
// main.go - Função main()
// Configura o token de autenticação a partir da variável de ambiente
expectedAuthToken = os.Getenv("EXPECTED_AUTH_TOKEN")
if expectedAuthToken == "" {
    logger.Error("Variável de ambiente EXPECTED_AUTH_TOKEN não definida. Defina um token de autenticação.")
    os.Exit(1)
}
// Remove qualquer quebra de linha ou espaço em branco do token
expectedAuthToken = strings.TrimSpace(expectedAuthToken)
logger.Info("Token de autenticação configurado com sucesso.")
```

Esta mudança garante que tokens com quebras de linha ou espaços em branco indesejados ainda funcionem corretamente.

## Desenvolvimento Local

### Compilar e Executar

Para compilar e executar a aplicação localmente:

```bash
# Clone o repositório
git clone https://github.com/nataliagranato/k8s-api-metrics.git
cd k8s-api-metrics

# Baixe as dependências
go mod download
go mod verify

# Compile o código
go build -o k8s-api-metrics .

# Configure o token de autenticação
export EXPECTED_AUTH_TOKEN="meuTokenSuperSeguro123!@#"

# Execute a aplicação (usa o kubeconfig do seu $HOME/.kube/config)
./k8s-api-metrics
```

### Testar Localmente

Em outro terminal, teste os endpoints da API:

```bash
# Teste o endpoint de health check (não requer autenticação)
curl http://localhost:8080/healthz

# Teste o endpoint de métricas JSON (requer autenticação)
curl -H 'Authorization: Bearer meuTokenSuperSeguro123!@#' http://localhost:8080/metrics

# Teste o endpoint de métricas Prometheus (requer autenticação)
curl -H 'Authorization: Bearer meuTokenSuperSeguro123!@#' http://localhost:8080/metrics-prometheus
```

## Docker

### Build da Imagem

Para construir a imagem Docker:

```bash
# Build da imagem
docker build -t nataliagranato/k8s-api-metrics:v1.0.1 .

# Verificar se a imagem foi criada
docker images | grep k8s-api-metrics
```

### Executar com Docker

Para executar a aplicação usando Docker:

```bash
# Executar o container Docker
docker run -p 8080:8080 -e EXPECTED_AUTH_TOKEN="meuTokenSuperSeguro123!@#" nataliagranato/k8s-api-metrics:v1.0.1

# Para execução em modo detached (background)
docker run -d -p 8080:8080 -e EXPECTED_AUTH_TOKEN="meuTokenSuperSeguro123!@#" nataliagranato/k8s-api-metrics:v1.0.1
```

Observação: Para que o container acesse o cluster Kubernetes, você precisará montar o arquivo kubeconfig ou executar o container dentro de um pod no cluster.

### Publicar Imagem no Registry

Para publicar a imagem no Docker Hub ou outro registry:

```bash
# Login no Docker Hub
docker login

# Push da imagem
docker push nataliagranato/k8s-api-metrics:v1.0.1
```

## Kubernetes

### Deploy com Helm

Para implantar a aplicação no Kubernetes usando Helm:

```bash
# Navegue até a pasta do projeto
cd k8s-api-metrics

# Crie um namespace para a aplicação
kubectl create namespace k8s-api-metrics

# Instale o chart Helm
helm install k8s-api-metrics ./charts/k8s-api-metrics -n k8s-api-metrics
```

Para personalizar a instalação, você pode editar o arquivo `charts/k8s-api-metrics/values.yaml` ou usar o parâmetro `--set`:

```bash
helm install k8s-api-metrics ./charts/k8s-api-metrics -n k8s-api-metrics \
  --set application.authToken="seuTokenPersonalizado" \
  --set replicaCount=2
```

### Acessar a API no Kubernetes

Após a instalação, você pode acessar a API usando port-forward:

```bash
# Port-forward para acessar o serviço localmente
kubectl port-forward svc/k8s-api-metrics-k8s-api-metrics 8080:8080 -n k8s-api-metrics

# Em outro terminal, acesse a API com o token
curl -H 'Authorization: Bearer meuTokenSuperSeguro123!@#' http://localhost:8080/metrics
```

Ou, se você configurou um Ingress, através do hostname configurado.

### Atualizar o Deployment

Para atualizar a aplicação para uma nova versão:

```bash
# Edite o arquivo values.yaml para atualizar a tag da imagem
# Em seguida, atualize o deployment
helm upgrade k8s-api-metrics ./charts/k8s-api-metrics -n k8s-api-metrics

# Verificar o status da atualização
kubectl get pods -n k8s-api-metrics
```

### Troubleshooting

Para diagnosticar problemas:

```bash
# Verificar os pods
kubectl get pods -n k8s-api-metrics

# Verificar os logs do pod
kubectl logs -n k8s-api-metrics $(kubectl get pods -n k8s-api-metrics -o jsonpath='{.items[0].metadata.name}')

# Verificar o Secret com o token
kubectl get secret -n k8s-api-metrics k8s-api-metrics-k8s-api-metrics-auth-token -o jsonpath='{.data.auth-token}' | base64 --decode

# Verificar o token dentro do pod
kubectl exec -it -n k8s-api-metrics $(kubectl get pods -n k8s-api-metrics -o jsonpath='{.items[0].metadata.name}') -- sh
echo "$EXPECTED_AUTH_TOKEN"
```

## Endpoints da API

A API expõe os seguintes endpoints:

- `/metrics` - Métricas em formato JSON (requer autenticação)
- `/metrics-prometheus` - Métricas em formato Prometheus (requer autenticação)
- `/healthz` - Endpoint de health check (não requer autenticação)

### Exemplos de Resposta

#### `/metrics` (JSON)

```json
{
  "nodeCount": 1,
  "podCount": 10,
  "deploymentCount": 3,
  "serviceCount": 3,
  "nodes": [
    {
      "name": "kind-control-plane",
      "status": "Ready",
      "allocatableCpu": "12",
      "allocatableMemory": "16053964Ki",
      "kubeletVersion": "v1.30.0",
      "osImage": "Debian GNU/Linux 12 (bookworm)",
      "labels": {
        "beta.kubernetes.io/arch": "amd64",
        "beta.kubernetes.io/os": "linux",
        "kubernetes.io/arch": "amd64",
        "kubernetes.io/hostname": "kind-control-plane",
        "kubernetes.io/os": "linux",
        "node-role.kubernetes.io/control-plane": ""
      }
    }
  ],
  "timestamp": "2025-05-27T23:42:58.553630851Z"
}
```

#### `/healthz` (Health Check)

```json
{
  "status": "ok",
  "timestamp": "2025-05-27T23:43:15Z"
}
```

## Autenticação

A API utiliza autenticação baseada em token usando o header HTTP `Authorization`. 

Exemplo:
```
Authorization: Bearer meuTokenSuperSeguro123!@#
```

O token esperado é configurado através da variável de ambiente `EXPECTED_AUTH_TOKEN` no container.

## Observações e Melhorias

- A partir da versão v1.0.1, a aplicação utiliza `strings.TrimSpace()` para remover quebras de linha ou espaços em branco indesejados no token de autenticação, evitando problemas comuns com tokens inválidos.
- Para ambientes de produção, recomenda-se usar um Secret existente (opção 2 na configuração do Helm) e não definir o token diretamente no values.yaml.
- A API está configurada para coletar métricas básicas do cluster. Para métricas mais avançadas, seria necessário integrar com o Metrics Server ou o Prometheus.

## Contribuindo

Para contribuir com o projeto:

1. Faça um fork do repositório
2. Crie uma branch para sua feature (`git checkout -b feature/nova-feature`)
3. Implemente suas alterações
4. Execute os testes (se houver)
5. Faça commit das alterações (`git commit -m 'Adiciona nova feature'`)
6. Faça push para a branch (`git push origin feature/nova-feature`)
7. Abra um Pull Request

### Validação e Testes

Para validar que suas alterações funcionam corretamente, siga estes passos:

#### 1. Teste Local com Go

```bash
# Execute a aplicação localmente
export EXPECTED_AUTH_TOKEN="meuTokenSuperSeguro123!@#"
go run main.go
```

Em outro terminal:
```bash
# Teste o endpoint de health
curl http://localhost:8080/healthz

# Teste o endpoint de métricas com autenticação
curl -H 'Authorization: Bearer meuTokenSuperSeguro123!@#' http://localhost:8080/metrics
```

#### 2. Teste com Docker

```bash
# Construa a imagem
docker build -t nataliagranato/k8s-api-metrics:latest .

# Execute o container
docker run -p 8080:8080 -e EXPECTED_AUTH_TOKEN="meuTokenSuperSeguro123!@#" nataliagranato/k8s-api-metrics:latest
```

#### 3. Teste com Kubernetes

```bash
# Deploy com Helm
helm install k8s-api-metrics ./charts/k8s-api-metrics -n k8s-api-metrics --create-namespace

# Verifique os pods
kubectl get pods -n k8s-api-metrics

# Port-forward
kubectl port-forward svc/k8s-api-metrics-k8s-api-metrics 8080:8080 -n k8s-api-metrics

# Teste a API
curl -H 'Authorization: Bearer meuTokenSuperSeguro123!@#' http://localhost:8080/metrics
```


### Convenções de Código

- Siga as [convenções de código Go](https://golang.org/doc/effective_go)
- Use `gofmt` para formatar seu código
- Adicione comentários explicativos para funções e estruturas complexas
- Mantenha a documentação atualizada

## Licença

Este projeto está licenciado sob [inserir licença aqui] - veja o arquivo LICENSE para detalhes.

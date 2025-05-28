# k8s-metrics-api Helm Chart

Este Helm chart instala a API de métricas Kubernetes, que fornece informações sobre o cluster Kubernetes através de endpoints RESTful protegidos por autenticação.

## Instalação

```bash

# Clone o repositório
git clone https://github.com/nataliagranato/k8s-api-metrics.git

cd k8s-api-metrics

# Instalando o chart
helm install k8s-api-metrics ./charts/k8s-metrics-api -n k8s-api-metrics --create-namespace
```

## Configuração de Autenticação

A API requer autenticação via token. Existem duas opções para configurar este token:

### Opção 1: Permitir que o chart crie o Secret (recomendado para desenvolvimento)

No arquivo `values.yaml`, defina:
```yaml
application:
  authToken: "seu-token-aqui"  # Substitua por um token seguro em produção
  createAuthSecret: true
  existingAuthSecretName: ""   # Deixe vazio quando createAuthSecret for true
```

## Acessando a API

Após a instalação, você pode acessar a API usando port-forward:

```bash
# Port-forward para acessar o serviço localmente
kubectl port-forward svc/k8s-api-metrics-k8s-metrics-api 8080:8080 -n k8s-api-metrics

# Em outro terminal, acesse a API com o token
# Use aspas simples para evitar problemas com caracteres especiais no token
curl -H 'Authorization: Bearer meuTokenSuperSeguro123!@#' http://localhost:8080/metrics

# Teste o endpoint de métricas Prometheus
curl -H 'Authorization: Bearer meuTokenSuperSeguro123!@#' http://localhost:8080/metrics-prometheus

# Teste o endpoint de health check (não requer autenticação)
curl http://localhost:8080/healthz
```

Para verificar o token usado pelo pod:

```bash
# Obter o nome do pod
POD_NAME=$(kubectl get pods -n k8s-api-metrics -o jsonpath='{.items[0].metadata.name}')

# Verificar o token configurado
kubectl exec -it $POD_NAME -n k8s-api-metrics -- sh -c 'echo "$EXPECTED_AUTH_TOKEN"'

# Verificar o token armazenado no Secret
kubectl get secret -n k8s-api-metrics k8s-api-metrics-k8s-metrics-api-auth-token -o jsonpath='{.data.auth-token}' | base64 --decode
```

## Upgrade e Atualizações

Para atualizar uma instalação existente do chart:

```bash
# Atualize o Helm chart para a versão mais recente
helm upgrade k8s-api-metrics ./charts/k8s-metrics-api -n k8s-api-metrics

# Para atualizar com valores personalizados
helm upgrade k8s-api-metrics ./charts/k8s-metrics-api -n k8s-api-metrics \
  --set image.tag=v1.0.1 \
  --set application.authToken="meuNovoTokenSuperSeguro"
```

### Mudanças na versão v1.0.1

A versão v1.0.1 inclui as seguintes melhorias:

1. Correção do problema de autenticação com tokens que continham quebras de linha
2. Uso de `strings.TrimSpace()` para limpar o token recebido via variável de ambiente
3. Melhor tratamento de erros e logs mais informativos

É altamente recomendado atualizar para esta versão se você estiver enfrentando problemas de autenticação.

## Endpoints disponíveis

- `/metrics` - Métricas em formato JSON (requer autenticação)
- `/metrics-prometheus` - Métricas em formato Prometheus (requer autenticação)
- `/healthz` - Endpoint de health check (não requer autenticação)

## Parâmetros

| Parâmetro                            | Descrição                                                | Valor Padrão                     |
| ------------------------------------ | -------------------------------------------------------- | -------------------------------- |
| `replicaCount`                       | Número de réplicas                                       | `1`                              |
| `image.repository`                   | Repositório da imagem Docker                             | `nataliagranato/k8s-metrics-api` |
| `image.tag`                          | Tag da imagem Docker                                     | `v1.0.1`                         |
| `image.pullPolicy`                   | Política de pull da imagem                               | `IfNotPresent`                   |
| `service.type`                       | Tipo de serviço Kubernetes                               | `ClusterIP`                      |
| `service.port`                       | Porta do serviço                                         | `8080`                           |
| `application.authToken`              | Token de autenticação quando createAuthSecret=true       | `"2iKpp86QEmbZnJ15z0XGSSrt"`     |
| `application.createAuthSecret`       | Se true, cria um Secret com o authToken                  | `true`                           |
| `application.existingAuthSecretName` | Nome do Secret existente (quando createAuthSecret=false) | `""`                             |
| `application.authSecretKey`          | Chave dentro do Secret que contém o token                | `"auth-token"`                   |
| `application.containerPort`          | Porta que a aplicação escuta dentro do container         | `8080`                           |
| `rbac.create`                        | Se true, cria recursos RBAC                              | `true`                           |

## Solução de Problemas

### Erro "token não fornecido" ou "token inválido"

Verifique se:

1. O Secret está sendo criado corretamente
2. O pod está recebendo a variável de ambiente EXPECTED_AUTH_TOKEN
3. O token que você está enviando na requisição é idêntico ao token armazenado no Secret

A aplicação na versão v1.0.1 usa `strings.TrimSpace()` para remover quaisquer quebras de linha ou espaços em branco no token, o que evita problemas comuns com tokens inválidos.

Você pode verificar o token dentro do pod com:
```bash
kubectl exec -it POD_NAME -n k8s-api-metrics -- sh
echo "$EXPECTED_AUTH_TOKEN"
```

### Erro de conexão ao acessar a API

Certifique-se de que:
1. O pod está rodando (kubectl get pods -n k8s-api-metrics)
2. O port-forward está ativo em um terminal separado
3. Você está usando o comando curl no formato correto

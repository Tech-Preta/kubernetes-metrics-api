#!/bin/bash

set -e
set -x

# Função para verificar se o Docker está funcionando
check_docker() {
  echo "Verificando conectividade com o Docker..."
  if ! sudo docker info > /dev/null 2>&1; then
    echo "ERRO: Não foi possível conectar ao Docker. Verifique se o socket está mapeado corretamente."
    return 1
  fi
  echo "Conectividade com Docker OK!"
  return 0
}

# Função para criar um cluster KinD
create_kind_cluster() {
  echo "Criando cluster KinD..."

  # Verificar se já existe um cluster com este nome
  if sudo kind get clusters 2>/dev/null | grep -q "devcontainer-cluster"; then
    echo "Cluster 'devcontainer-cluster' já existe."
    return 0
  fi

  # Configuração do cluster KinD com mapeamento de portas
  cat > /tmp/kind-config.yaml << EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 8081
    hostPort: 8081
    protocol: TCP
  - containerPort: 443
    hostPort: 8443
    protocol: TCP
  - containerPort: 6443
    hostPort: 16443    # novo hostPort para evitar conflito
    protocol: TCP
EOF

  sudo kind create cluster --name devcontainer-cluster --config /tmp/kind-config.yaml

  echo "Configurando o contexto do kubectl..."
  mkdir -p $HOME/.kube
  sudo kind get kubeconfig --name devcontainer-cluster > $HOME/.kube/config

  # Corrige o endpoint do servidor para localhost:16443 e nomes de contexto/cluster/user
  sed -i 's|server: https://127.0.0.1:[0-9]*|server: https://127.0.0.1:16443|' $HOME/.kube/config
  sed -i 's|name: kind-.*|name: kind-devcontainer-cluster|g' $HOME/.kube/config
  sed -i 's|cluster: kind-.*|cluster: kind-devcontainer-cluster|g' $HOME/.kube/config
  sed -i 's|user: kind-.*|user: kind-devcontainer-cluster|g' $HOME/.kube/config
  sed -i 's|current-context: kind-.*|current-context: kind-devcontainer-cluster|g' $HOME/.kube/config

  sudo chown vscode:vscode $HOME/.kube/config

  return 0
}

# Instalar dependências do projeto (opcional)
install_project_deps() {
  echo "Instalando dependências adicionais do projeto..."

  # Configurar mise para ativação automática
  echo "Configurando mise..."
  if command -v mise &> /dev/null; then
    echo "✓ Mise encontrado, configurando ativação automática..."

    # Adicionar ativação do mise ao .bashrc se não estiver presente
    if ! grep -q "mise activate" ~/.bashrc; then
      echo 'eval "$(mise activate bash)"' >> ~/.bashrc
      echo "  → Ativação automática do mise adicionada ao ~/.bashrc"
    fi

    # Ativar mise na sessão atual
    eval "$(mise activate bash)"

    # Instalar ferramentas definidas no mise.toml
    if [ -f "mise.toml" ]; then
      mise install
      echo "  → Ferramentas do mise.toml instaladas"
    fi

    echo "✓ Mise configurado com sucesso"
  else
    echo "✗ Mise não encontrado"
  fi

  # Verificar se as ferramentas DevOps estão instaladas e funcionando
  echo "Verificando ferramentas DevOps..."

  # Verificar Terraform
  if command -v terraform &> /dev/null; then
    echo "✓ Terraform: $(terraform version -json | jq -r '.terraform_version')"
  else
    echo "✗ Terraform não encontrado"
  fi

  # Verificar OpenTofu
  if command -v tofu &> /dev/null; then
    echo "✓ OpenTofu: $(tofu version | head -1)"
  else
    echo "✗ OpenTofu não encontrado"
  fi

  # Verificar terraform-docs
  if command -v terraform-docs &> /dev/null; then
    echo "✓ terraform-docs: $(terraform-docs version)"
  else
    echo "✗ terraform-docs não encontrado"
  fi

  # Verificar Terragrunt
  if command -v terragrunt &> /dev/null; then
    echo "✓ Terragrunt: $(terragrunt --version | head -1)"
  else
    echo "✗ Terragrunt não encontrado"
  fi

  # Verificar Golang
  if command -v go &> /dev/null; then
    echo "✓ Golang: $(go version)"
  else
    echo "✗ Golang não encontrado"
  fi

  # Verificar AWS CLI
  if command -v aws &> /dev/null; then
    echo "✓ AWS CLI: $(aws --version)"
  else
    echo "✗ AWS CLI não encontrado"
  fi

  # Verificar Azure CLI
  if command -v az &> /dev/null; then
    echo "✓ Azure CLI: $(az version --output tsv --query '\"azure-cli\"')"
  else
    echo "✗ Azure CLI não encontrado"
  fi

  # Verificar Google Cloud SDK
  if command -v gcloud &> /dev/null; then
    echo "✓ Google Cloud SDK: $(gcloud version --format='value(Google Cloud SDK)' 2>/dev/null | head -1)"
  else
    echo "✗ Google Cloud SDK não encontrado"
  fi

  # Verificar GitHub CLI
  if command -v gh &> /dev/null; then
    echo "✓ GitHub CLI: $(gh version | head -1)"
  else
    echo "✗ GitHub CLI não encontrado"
  fi

  # Verificar Ansible
  if command -v ansible &> /dev/null; then
    echo "✓ Ansible: $(ansible --version | head -1)"
  else
    echo "✗ Ansible não encontrado"
  fi

  # Verificar Packer
  if command -v packer &> /dev/null; then
    echo "✓ Packer: $(packer version)"
  else
    echo "✗ Packer não encontrado"
  fi

  # Verificar Pulumi
  if command -v pulumi &> /dev/null; then
    echo "✓ Pulumi: $(pulumi version)"
  else
    echo "✗ Pulumi não encontrado"
  fi

  # Verificar outras ferramentas essenciais
  for tool in kubectl helm k9s kind stern yq jq curl httpie make; do
    if command -v $tool &> /dev/null; then
      echo "✓ $tool: disponível"
    else
      echo "✗ $tool não encontrado"
    fi
  done
}

# Execução principal
main() {
  echo "Iniciando configuração do ambiente de desenvolvimento..."

  if check_docker; then
    create_kind_cluster
    install_project_deps
    echo "Ambiente de desenvolvimento configurado com sucesso!"
  else
    echo "Falha ao configurar o ambiente de desenvolvimento."
    exit 1
  fi
}

# Executar o script principal
main

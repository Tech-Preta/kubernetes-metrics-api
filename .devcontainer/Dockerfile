FROM mcr.microsoft.com/vscode/devcontainers/base:ubuntu-22.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg \
    lsb-release \
    software-properties-common \
    git \
    unzip \
    jq \
    sudo \
    make \
    python3 \
    python3-pip \
    wget \
    && rm -rf /var/lib/apt/lists/*

# Docker CLI
RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" > /etc/apt/sources.list.d/docker.list \
    && apt-get update && apt-get install -y docker-ce-cli docker-buildx-plugin docker-compose-plugin \
    && rm -rf /var/lib/apt/lists/*

# kubectl
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" \
    && chmod +x kubectl \
    && mv kubectl /usr/local/bin/

# KinD
RUN curl -Lo /usr/local/bin/kind https://kind.sigs.k8s.io/dl/v0.21.0/kind-linux-amd64 \
    && chmod +x /usr/local/bin/kind

# Helm
RUN curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Golang
RUN curl -fsSL https://go.dev/dl/go1.24.5.linux-amd64.tar.gz -o go.tar.gz \
    && tar -C /usr/local -xzf go.tar.gz \
    && rm go.tar.gz
ENV PATH="/usr/local/go/bin:${PATH}"

# Terraform
RUN curl -fsSL https://apt.releases.hashicorp.com/gpg | gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg \
    && echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" > /etc/apt/sources.list.d/hashicorp.list \
    && apt-get update && apt-get install -y terraform packer vault \
    && rm -rf /var/lib/apt/lists/*

# OpenTofu
RUN curl -fsSL https://get.opentofu.org/install-opentofu.sh -o install-opentofu.sh \
    && chmod +x install-opentofu.sh \
    && ./install-opentofu.sh --install-method deb \
    && rm install-opentofu.sh

# terraform-docs
RUN TERRAFORM_DOCS_VERSION=$(curl -s https://api.github.com/repos/terraform-docs/terraform-docs/releases/latest | grep tag_name | cut -d '"' -f 4) \
    && curl -Lo terraform-docs.tar.gz "https://github.com/terraform-docs/terraform-docs/releases/latest/download/terraform-docs-${TERRAFORM_DOCS_VERSION}-linux-amd64.tar.gz" \
    && tar -xzf terraform-docs.tar.gz \
    && chmod +x terraform-docs \
    && mv terraform-docs /usr/local/bin/ \
    && rm terraform-docs.tar.gz

# Terragrunt
RUN curl -Lo /usr/local/bin/terragrunt https://github.com/gruntwork-io/terragrunt/releases/latest/download/terragrunt_linux_amd64 \
    && chmod +x /usr/local/bin/terragrunt

# AWS CLI v2
RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" \
    && unzip awscliv2.zip \
    && ./aws/install \
    && rm -rf awscliv2.zip aws

# Python-based tools
RUN pip3 install --no-cache-dir \
    ansible \
    httpie \
    pre-commit

# Azure CLI
RUN curl -sL https://aka.ms/InstallAzureCLIDeb | bash

# Google Cloud SDK
RUN echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" > /etc/apt/sources.list.d/google-cloud-sdk.list \
    && curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | gpg --dearmor -o /usr/share/keyrings/cloud.google.gpg \
    && apt-get update && apt-get install -y google-cloud-cli \
    && rm -rf /var/lib/apt/lists/*

# GitHub CLI
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | gpg --dearmor -o /usr/share/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" > /etc/apt/sources.list.d/github-cli.list \
    && apt-get update && apt-get install -y gh \
    && rm -rf /var/lib/apt/lists/*

# Puppet
RUN wget -O /tmp/puppetlabs-release-pc1-jammy.deb https://apt.puppetlabs.com/puppetlabs-release-pc1-jammy.deb \
    && dpkg -i /tmp/puppetlabs-release-pc1-jammy.deb \
    && apt-get update && apt-get install -y puppet-agent \
    && rm -f /tmp/puppetlabs-release-pc1-jammy.deb \
    && rm -rf /var/lib/apt/lists/*
ENV PATH="/opt/puppetlabs/bin:${PATH}"

# Pulumi
RUN curl -fsSL https://get.pulumi.com | sh
ENV PATH="/root/.pulumi/bin:${PATH}"

# Chef Workstation
RUN curl -L https://omnitruck.chef.io/install.sh | bash -s -- -P chef-workstation

# yq
RUN YQ_VERSION=$(curl -s https://api.github.com/repos/mikefarah/yq/releases/latest | grep tag_name | cut -d '"' -f 4 | head -1) \
    && curl -Lo /usr/local/bin/yq "https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64" \
    && chmod +x /usr/local/bin/yq

# stern
RUN STERN_VERSION=$(curl -s https://api.github.com/repos/stern/stern/releases/latest | grep tag_name | cut -d '"' -f 4 | head -1) \
    && curl -Lo /usr/local/bin/stern "https://github.com/stern/stern/releases/latest/download/stern_linux_amd64" \
    && chmod +x /usr/local/bin/stern

# tfsec
RUN TFSEC_VERSION=$(curl -s https://api.github.com/repos/aquasecurity/tfsec/releases/latest | grep tag_name | cut -d '"' -f 4 | head -1) \
    && curl -Lo /usr/local/bin/tfsec "https://github.com/aquasecurity/tfsec/releases/latest/download/tfsec-linux-amd64" \
    && chmod +x /usr/local/bin/tfsec

# kubescape
RUN curl -s https://raw.githubusercontent.com/kubescape/kubescape/master/install.sh | bash || \
    (curl -Lo /usr/local/bin/kubescape "https://github.com/kubescape/kubescape/releases/latest/download/kubescape-ubuntu-latest" \
    && chmod +x /usr/local/bin/kubescape)

# kube-bench
RUN KUBE_BENCH_VERSION=$(curl -s https://api.github.com/repos/aquasecurity/kube-bench/releases/latest | grep tag_name | cut -d '"' -f 4 | head -1) \
    && curl -Lo kube-bench.tar.gz "https://github.com/aquasecurity/kube-bench/releases/latest/download/kube-bench_${KUBE_BENCH_VERSION#v}_linux_amd64.tar.gz" \
    && tar -xzf kube-bench.tar.gz \
    && mv kube-bench /usr/local/bin/ \
    && rm kube-bench.tar.gz

# kubectx and kubens
RUN KUBECTX_VERSION=$(curl -s https://api.github.com/repos/ahmetb/kubectx/releases/latest | grep tag_name | cut -d '"' -f 4 | head -1) \
    && curl -Lo /usr/local/bin/kubectx "https://github.com/ahmetb/kubectx/releases/latest/download/kubectx" \
    && curl -Lo /usr/local/bin/kubens "https://github.com/ahmetb/kubectx/releases/latest/download/kubens" \
    && chmod +x /usr/local/bin/kubectx /usr/local/bin/kubens

# Kustomize (standalone)
RUN KUSTOMIZE_VERSION=$(curl -s https://api.github.com/repos/kubernetes-sigs/kustomize/releases/latest | grep tag_name | grep kustomize | head -1 | cut -d '"' -f 4) \
    && curl -Lo kustomize.tar.gz "https://github.com/kubernetes-sigs/kustomize/releases/latest/download/kustomize_${KUSTOMIZE_VERSION#kustomize/v}_linux_amd64.tar.gz" \
    && tar -xzf kustomize.tar.gz \
    && mv kustomize /usr/local/bin/ \
    && rm kustomize.tar.gz

# Sops
RUN SOPS_VERSION=$(curl -s https://api.github.com/repos/getsops/sops/releases/latest | grep tag_name | cut -d '"' -f 4 | head -1) \
    && curl -Lo /usr/local/bin/sops "https://github.com/getsops/sops/releases/latest/download/sops-${SOPS_VERSION}.linux.amd64" \
    && chmod +x /usr/local/bin/sops

# Minikube
RUN curl -Lo /usr/local/bin/minikube "https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64" \
    && chmod +x /usr/local/bin/minikube

# Skaffold
RUN curl -Lo /usr/local/bin/skaffold "https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64" \
    && chmod +x /usr/local/bin/skaffold

# Task
RUN TASK_VERSION=$(curl -s https://api.github.com/repos/go-task/task/releases/latest | grep tag_name | cut -d '"' -f 4 | head -1) \
    && curl -Lo task.tar.gz "https://github.com/go-task/task/releases/latest/download/task_linux_amd64.tar.gz" \
    && tar -xzf task.tar.gz \
    && mv task /usr/local/bin/ \
    && rm task.tar.gz

# promtool (from Prometheus)
RUN PROMETHEUS_VERSION=$(curl -s https://api.github.com/repos/prometheus/prometheus/releases/latest | grep tag_name | cut -d '"' -f 4 | head -1) \
    && curl -Lo prometheus.tar.gz "https://github.com/prometheus/prometheus/releases/latest/download/prometheus-${PROMETHEUS_VERSION#v}.linux-amd64.tar.gz" \
    && tar -xzf prometheus.tar.gz \
    && mv prometheus-${PROMETHEUS_VERSION#v}.linux-amd64/promtool /usr/local/bin/ \
    && rm -rf prometheus.tar.gz prometheus-${PROMETHEUS_VERSION#v}.linux-amd64

# Fix K9s installation
RUN K9S_VERSION=$(curl -s https://api.github.com/repos/derailed/k9s/releases/latest | grep tag_name | cut -d '"' -f 4 | head -1) \
    && curl -Lo k9s.tar.gz "https://github.com/derailed/k9s/releases/latest/download/k9s_Linux_amd64.tar.gz" \
    && tar -xzf k9s.tar.gz \
    && mv k9s /usr/local/bin/ \
    && rm k9s.tar.gz LICENSE README.md

# Permissões e sudo sem senha para vscode
RUN groupadd -f docker \
    && usermod -aG docker vscode \
    && mkdir -p /home/vscode/.kube /home/vscode/.kind \
    && chown -R vscode:vscode /home/vscode/.kube /home/vscode/.kind \
    && echo "vscode ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers.d/99-vscode

USER vscode

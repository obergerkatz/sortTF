# Multiple providers with aliases
terraform {
  required_version = ">= 1.5"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.0"
    }
    datadog = {
      source  = "datadog/datadog"
      version = "~> 3.0"
    }
  }
}

# AWS Providers - Multiple regions
provider "aws" {
  region = "us-east-1"
  alias  = "primary"

  default_tags {
    tags = {
      Environment = "production"
      ManagedBy   = "Terraform"
      Region      = "us-east-1"
    }
  }
}

provider "aws" {
  region = "us-west-2"
  alias  = "secondary"

  default_tags {
    tags = {
      Environment = "production"
      ManagedBy   = "Terraform"
      Region      = "us-west-2"
    }
  }
}

provider "aws" {
  region = "eu-west-1"
  alias  = "europe"

  default_tags {
    tags = {
      Environment = "production"
      ManagedBy   = "Terraform"
      Region      = "eu-west-1"
    }
  }
}

# AWS Provider with assume role
provider "aws" {
  region = "us-east-1"
  alias  = "cross_account"

  assume_role {
    role_arn     = "arn:aws:iam::123456789012:role/TerraformExecutionRole"
    session_name = "terraform-cross-account"
    external_id  = "unique-external-id"
  }
}

# Google Cloud Provider
provider "google" {
  project = "my-gcp-project"
  region  = "us-central1"
  alias   = "main"
}

provider "google" {
  project = "my-gcp-project"
  region  = "europe-west1"
  alias   = "europe"
}

# Azure Provider
provider "azurerm" {
  features {
    resource_group {
      prevent_deletion_if_contains_resources = true
    }
    key_vault {
      purge_soft_delete_on_destroy    = false
      recover_soft_deleted_key_vaults = true
    }
  }

  subscription_id = var.azure_subscription_id
  tenant_id       = var.azure_tenant_id
}

# Kubernetes Provider - using AWS EKS
provider "kubernetes" {
  host                   = data.aws_eks_cluster.primary.endpoint
  cluster_ca_certificate = base64decode(data.aws_eks_cluster.primary.certificate_authority[0].data)
  token                  = data.aws_eks_cluster_auth.primary.token

  alias = "eks_primary"
}

provider "kubernetes" {
  host                   = data.aws_eks_cluster.secondary.endpoint
  cluster_ca_certificate = base64decode(data.aws_eks_cluster.secondary.certificate_authority[0].data)
  token                  = data.aws_eks_cluster_auth.secondary.token

  alias = "eks_secondary"
}

# Helm Provider
provider "helm" {
  kubernetes {
    host                   = data.aws_eks_cluster.primary.endpoint
    cluster_ca_certificate = base64decode(data.aws_eks_cluster.primary.certificate_authority[0].data)
    token                  = data.aws_eks_cluster_auth.primary.token
  }

  alias = "eks_primary"
}

# Datadog Provider
provider "datadog" {
  api_key = var.datadog_api_key
  app_key = var.datadog_app_key
  api_url = "https://api.datadoghq.com/"
}

# Variables
variable "azure_subscription_id" {
  description = "Azure subscription ID"
  type        = string
}

variable "azure_tenant_id" {
  description = "Azure tenant ID"
  type        = string
}

variable "datadog_api_key" {
  description = "Datadog API key"
  type        = string
  sensitive   = true
}

variable "datadog_app_key" {
  description = "Datadog application key"
  type        = string
  sensitive   = true
}

# Data sources for EKS clusters
data "aws_eks_cluster" "primary" {
  provider = aws.primary
  name     = "my-cluster-us-east-1"
}

data "aws_eks_cluster_auth" "primary" {
  provider = aws.primary
  name     = "my-cluster-us-east-1"
}

data "aws_eks_cluster" "secondary" {
  provider = aws.secondary
  name     = "my-cluster-us-west-2"
}

data "aws_eks_cluster_auth" "secondary" {
  provider = aws.secondary
  name     = "my-cluster-us-west-2"
}

# Resources using different providers
resource "aws_s3_bucket" "primary_logs" {
  provider = aws.primary
  bucket   = "logs-us-east-1"
}

resource "aws_s3_bucket" "secondary_logs" {
  provider = aws.secondary
  bucket   = "logs-us-west-2"
}

resource "aws_s3_bucket" "europe_logs" {
  provider = aws.europe
  bucket   = "logs-eu-west-1"
}

resource "google_storage_bucket" "gcp_logs" {
  provider = google.main
  name     = "logs-gcp-us-central1"
  location = "US"
}

resource "azurerm_resource_group" "main" {
  provider = azurerm
  name     = "rg-main"
  location = "East US"
}

resource "kubernetes_namespace" "monitoring" {
  provider = kubernetes.eks_primary

  metadata {
    name = "monitoring"
  }
}

resource "helm_release" "prometheus" {
  provider = helm.eks_primary

  name       = "prometheus"
  repository = "https://prometheus-community.github.io/helm-charts"
  chart      = "kube-prometheus-stack"
  namespace  = kubernetes_namespace.monitoring.metadata[0].name

  values = [
    file("${path.module}/helm-values/prometheus.yaml")
  ]
}

resource "datadog_monitor" "high_cpu" {
  provider = datadog

  name    = "High CPU Usage"
  type    = "metric alert"
  message = "CPU usage is high @pagerduty"

  query = "avg(last_5m):avg:system.cpu.user{*} by {host} > 90"

  monitor_thresholds {
    critical = 90
    warning  = 80
  }
}

# Outputs from different providers
output "aws_primary_bucket" {
  description = "AWS primary bucket name"
  value       = aws_s3_bucket.primary_logs.id
}

output "gcp_bucket" {
  description = "GCP bucket name"
  value       = google_storage_bucket.gcp_logs.name
}

output "azure_resource_group" {
  description = "Azure resource group name"
  value       = azurerm_resource_group.main.name
}

output "kubernetes_namespace" {
  description = "Kubernetes namespace"
  value       = kubernetes_namespace.monitoring.metadata[0].name
}

# Large Kubernetes deployment configuration
terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
  }
}

provider "kubernetes" {
  config_path = "~/.kube/config"
}

variable "namespace" {
  description = "Kubernetes namespace"
  type        = string
  default     = "production"
}

variable "app_name" {
  description = "Application name"
  type        = string
  default     = "myapp"
}

variable "replicas" {
  description = "Number of replicas"
  type        = number
  default     = 3
}

variable "image_tag" {
  description = "Docker image tag"
  type        = string
  default     = "latest"
}

locals {
  common_labels = {
    app         = var.app_name
    environment = var.namespace
    managed_by  = "terraform"
  }

  backend_labels = merge(local.common_labels, {
    component = "backend"
    tier      = "api"
  })

  frontend_labels = merge(local.common_labels, {
    component = "frontend"
    tier      = "web"
  })

  database_labels = merge(local.common_labels, {
    component = "database"
    tier      = "data"
  })

  cache_labels = merge(local.common_labels, {
    component = "cache"
    tier      = "data"
  })
}

resource "kubernetes_namespace" "app" {
  metadata {
    name = var.namespace
    labels = local.common_labels
  }
}

resource "kubernetes_config_map" "app_config" {
  metadata {
    name      = "${var.app_name}-config"
    namespace = kubernetes_namespace.app.metadata[0].name
    labels    = local.common_labels
  }

  data = {
    APP_NAME         = var.app_name
    APP_ENV          = var.namespace
    LOG_LEVEL        = "info"
    DATABASE_HOST    = "postgres.${var.namespace}.svc.cluster.local"
    DATABASE_PORT    = "5432"
    CACHE_HOST       = "redis.${var.namespace}.svc.cluster.local"
    CACHE_PORT       = "6379"
    API_TIMEOUT      = "30s"
    MAX_CONNECTIONS  = "100"
    ENABLE_METRICS   = "true"
    METRICS_PORT     = "9090"
  }
}

resource "kubernetes_secret" "app_secrets" {
  metadata {
    name      = "${var.app_name}-secrets"
    namespace = kubernetes_namespace.app.metadata[0].name
    labels    = local.common_labels
  }

  type = "Opaque"

  data = {
    DATABASE_PASSWORD = base64encode("super-secret-password")
    API_KEY           = base64encode("api-key-12345")
    JWT_SECRET        = base64encode("jwt-secret-key")
  }
}

resource "kubernetes_deployment" "backend" {
  metadata {
    name      = "${var.app_name}-backend"
    namespace = kubernetes_namespace.app.metadata[0].name
    labels    = local.backend_labels
  }

  spec {
    replicas = var.replicas

    selector {
      match_labels = {
        app       = var.app_name
        component = "backend"
      }
    }

    template {
      metadata {
        labels = local.backend_labels
        annotations = {
          "prometheus.io/scrape" = "true"
          "prometheus.io/port"   = "9090"
          "prometheus.io/path"   = "/metrics"
        }
      }

      spec {
        service_account_name = kubernetes_service_account.app.metadata[0].name

        container {
          name              = "backend"
          image             = "myregistry/backend:${var.image_tag}"
          image_pull_policy = "Always"

          port {
            name           = "http"
            container_port = 8080
            protocol       = "TCP"
          }

          port {
            name           = "metrics"
            container_port = 9090
            protocol       = "TCP"
          }

          env {
            name = "DATABASE_PASSWORD"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.app_secrets.metadata[0].name
                key  = "DATABASE_PASSWORD"
              }
            }
          }

          env {
            name = "API_KEY"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.app_secrets.metadata[0].name
                key  = "API_KEY"
              }
            }
          }

          env_from {
            config_map_ref {
              name = kubernetes_config_map.app_config.metadata[0].name
            }
          }

          resources {
            requests = {
              cpu    = "100m"
              memory = "128Mi"
            }
            limits = {
              cpu    = "500m"
              memory = "512Mi"
            }
          }

          liveness_probe {
            http_get {
              path = "/health"
              port = 8080
            }
            initial_delay_seconds = 30
            period_seconds        = 10
            timeout_seconds       = 5
            failure_threshold     = 3
          }

          readiness_probe {
            http_get {
              path = "/ready"
              port = 8080
            }
            initial_delay_seconds = 10
            period_seconds        = 5
            timeout_seconds       = 3
            failure_threshold     = 3
          }

          volume_mount {
            name       = "config"
            mount_path = "/app/config"
            read_only  = true
          }

          volume_mount {
            name       = "cache"
            mount_path = "/app/cache"
          }
        }

        volume {
          name = "config"
          config_map {
            name = kubernetes_config_map.app_config.metadata[0].name
          }
        }

        volume {
          name = "cache"
          empty_dir {}
        }

        affinity {
          pod_anti_affinity {
            preferred_during_scheduling_ignored_during_execution {
              weight = 100
              pod_affinity_term {
                label_selector {
                  match_expressions {
                    key      = "app"
                    operator = "In"
                    values   = [var.app_name]
                  }
                }
                topology_key = "kubernetes.io/hostname"
              }
            }
          }
        }
      }
    }

    strategy {
      type = "RollingUpdate"
      rolling_update {
        max_surge       = "25%"
        max_unavailable = "25%"
      }
    }
  }
}

resource "kubernetes_service" "backend" {
  metadata {
    name      = "${var.app_name}-backend"
    namespace = kubernetes_namespace.app.metadata[0].name
    labels    = local.backend_labels
  }

  spec {
    selector = {
      app       = var.app_name
      component = "backend"
    }

    port {
      name        = "http"
      port        = 80
      target_port = 8080
      protocol    = "TCP"
    }

    port {
      name        = "metrics"
      port        = 9090
      target_port = 9090
      protocol    = "TCP"
    }

    type = "ClusterIP"
  }
}

resource "kubernetes_horizontal_pod_autoscaler" "backend" {
  metadata {
    name      = "${var.app_name}-backend-hpa"
    namespace = kubernetes_namespace.app.metadata[0].name
  }

  spec {
    scale_target_ref {
      api_version = "apps/v1"
      kind        = "Deployment"
      name        = kubernetes_deployment.backend.metadata[0].name
    }

    min_replicas = 2
    max_replicas = 10

    metric {
      type = "Resource"
      resource {
        name = "cpu"
        target {
          type                = "Utilization"
          average_utilization = 70
        }
      }
    }

    metric {
      type = "Resource"
      resource {
        name = "memory"
        target {
          type                = "Utilization"
          average_utilization = 80
        }
      }
    }
  }
}

resource "kubernetes_service_account" "app" {
  metadata {
    name      = var.app_name
    namespace = kubernetes_namespace.app.metadata[0].name
    labels    = local.common_labels
  }
}

resource "kubernetes_role" "app" {
  metadata {
    name      = var.app_name
    namespace = kubernetes_namespace.app.metadata[0].name
  }

  rule {
    api_groups = [""]
    resources  = ["configmaps", "secrets"]
    verbs      = ["get", "list", "watch"]
  }

  rule {
    api_groups = [""]
    resources  = ["pods"]
    verbs      = ["get", "list"]
  }
}

resource "kubernetes_role_binding" "app" {
  metadata {
    name      = var.app_name
    namespace = kubernetes_namespace.app.metadata[0].name
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "Role"
    name      = kubernetes_role.app.metadata[0].name
  }

  subject {
    kind      = "ServiceAccount"
    name      = kubernetes_service_account.app.metadata[0].name
    namespace = kubernetes_namespace.app.metadata[0].name
  }
}

resource "kubernetes_ingress_v1" "app" {
  metadata {
    name      = var.app_name
    namespace = kubernetes_namespace.app.metadata[0].name
    annotations = {
      "kubernetes.io/ingress.class"                = "nginx"
      "cert-manager.io/cluster-issuer"             = "letsencrypt-prod"
      "nginx.ingress.kubernetes.io/ssl-redirect"   = "true"
      "nginx.ingress.kubernetes.io/force-ssl-redirect" = "true"
    }
  }

  spec {
    tls {
      hosts       = ["${var.app_name}.example.com"]
      secret_name = "${var.app_name}-tls"
    }

    rule {
      host = "${var.app_name}.example.com"
      http {
        path {
          path      = "/"
          path_type = "Prefix"
          backend {
            service {
              name = kubernetes_service.backend.metadata[0].name
              port {
                number = 80
              }
            }
          }
        }
      }
    }
  }
}

output "namespace" {
  description = "Kubernetes namespace"
  value       = kubernetes_namespace.app.metadata[0].name
}

output "backend_service" {
  description = "Backend service name"
  value       = kubernetes_service.backend.metadata[0].name
}

output "ingress_hostname" {
  description = "Ingress hostname"
  value       = "${var.app_name}.example.com"
}

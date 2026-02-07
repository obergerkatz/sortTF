# Lifecycle blocks and meta-arguments edge cases
resource "aws_instance" "all_lifecycle_options" {
  ami           = "ami-12345678"
  instance_type = "t3.micro"

  # All lifecycle options
  lifecycle {
    create_before_destroy = true
    prevent_destroy       = false
    ignore_changes = [
      tags,
      user_data,
      ami,
    ]
    replace_triggered_by = [
      null_resource.example.id
    ]
  }

  # Meta-arguments
  depends_on = [
    aws_security_group.example,
    aws_subnet.example,
  ]

  count = var.instance_count

  tags = {
    Name = "instance-${count.index}"
  }
}

resource "aws_instance" "for_each_example" {
  for_each = var.instances

  ami           = each.value.ami
  instance_type = each.value.type

  lifecycle {
    create_before_destroy = true
  }

  tags = {
    Name = each.key
    Type = each.value.type
  }
}

resource "null_resource" "depends_on_example" {
  # Complex depends_on with many resources
  depends_on = [
    aws_instance.all_lifecycle_options,
    aws_instance.for_each_example,
    aws_security_group.example,
    aws_subnet.example,
    aws_vpc.example,
    aws_route_table.example,
    aws_internet_gateway.example,
  ]

  provisioner "local-exec" {
    command = "echo 'All dependencies ready'"
  }
}

# Test sorting with provider meta-argument
resource "aws_instance" "provider_example" {
  provider = aws.alternate

  ami           = "ami-12345678"
  instance_type = "t3.micro"
}

# Test multiple lifecycle blocks scenario (one resource, one nested)
resource "kubernetes_deployment" "app" {
  metadata {
    name = "app"
  }

  spec {
    replicas = 3

    selector {
      match_labels = {
        app = "myapp"
      }
    }

    template {
      metadata {
        labels = {
          app = "myapp"
        }
      }

      spec {
        container {
          name  = "app"
          image = "nginx:latest"

          lifecycle {
            pre_stop {
              exec {
                command = ["/bin/sh", "-c", "sleep 30"]
              }
            }

            post_start {
              exec {
                command = ["/bin/sh", "-c", "echo Started"]
              }
            }
          }
        }
      }
    }
  }

  lifecycle {
    ignore_changes = [
      spec[0].replicas,
    ]
  }
}

# Moved block edge case
moved {
  from = aws_instance.old_name
  to   = aws_instance.new_name
}

moved {
  from = module.old_module.aws_instance.server
  to   = module.new_module.aws_instance.server
}

# Import block
import {
  to = aws_instance.imported
  id = "i-1234567890abcdef0"
}

# Check block (Terraform 1.5+)
check "health_check" {
  data "http" "example" {
    url = "https://example.com/health"
  }

  assert {
    condition     = data.http.example.status_code == 200
    error_message = "Health check failed"
  }
}

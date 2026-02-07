# Remote state configuration with multiple backend types
terraform {
  required_version = ">= 1.5"

  # S3 backend configuration
  backend "s3" {
    bucket         = "my-terraform-state"
    key            = "prod/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-state-lock"
    kms_key_id     = "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"

    # Workspace configuration
    workspace_key_prefix = "workspaces"

    # Additional S3 options
    acl                  = "private"
    skip_credentials_validation = false
    skip_metadata_api_check     = false
    skip_region_validation      = false
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.5"
    }
    null = {
      source  = "hashicorp/null"
      version = "~> 3.2"
    }
  }
}

# Reading remote state from other workspaces
data "terraform_remote_state" "network" {
  backend = "s3"
  config = {
    bucket = "my-terraform-state"
    key    = "network/terraform.tfstate"
    region = "us-east-1"
  }
}

data "terraform_remote_state" "security" {
  backend = "s3"
  config = {
    bucket = "my-terraform-state"
    key    = "security/terraform.tfstate"
    region = "us-east-1"
  }
}

data "terraform_remote_state" "database" {
  backend = "s3"
  config = {
    bucket = "my-terraform-state"
    key    = "database/terraform.tfstate"
    region = "us-east-1"
  }
}

# Using values from remote states
locals {
  vpc_id                = data.terraform_remote_state.network.outputs.vpc_id
  private_subnet_ids    = data.terraform_remote_state.network.outputs.private_subnet_ids
  public_subnet_ids     = data.terraform_remote_state.network.outputs.public_subnet_ids
  security_group_ids    = data.terraform_remote_state.security.outputs.security_group_ids
  database_endpoint     = data.terraform_remote_state.database.outputs.endpoint
  database_port         = data.terraform_remote_state.database.outputs.port
  database_name         = data.terraform_remote_state.database.outputs.database_name
}

# Resources using remote state values
resource "aws_instance" "app" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = "t3.medium"
  subnet_id              = local.private_subnet_ids[0]
  vpc_security_group_ids = local.security_group_ids

  user_data = templatefile("${path.module}/templates/app_config.sh", {
    db_endpoint = local.database_endpoint
    db_port     = local.database_port
    db_name     = local.database_name
  })

  tags = {
    Name = "app-server"
  }
}

data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"]

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }
}

# Export outputs for other states to consume
output "app_instance_id" {
  description = "Application instance ID"
  value       = aws_instance.app.id
}

output "app_private_ip" {
  description = "Application private IP"
  value       = aws_instance.app.private_ip
}

output "app_security_group_id" {
  description = "Application security group ID"
  value       = local.security_group_ids[0]
}

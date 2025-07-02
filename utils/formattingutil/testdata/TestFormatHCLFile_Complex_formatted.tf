# Complex Terraform configuration with poor formatting
terraform {
  required_version = ">= 1.0"
  backend "s3" {
    bucket  = "terraform-state-bucket"
    key     = "prod/terraform.tfstate"
    region  = "us-west-2"
    encrypt = true
  }
}

provider "aws" {
  region = var.region

  default_tags {
    tags = {
      Environment = "production"
      Project     = "infrastructure"
    }
  }
}

variable "region" {
  type        = string
  default     = "us-west-2"
  description = "AWS region"
}

locals {
  common_tags = {
    Owner       = "DevOps Team"
    Environment = "production"
    CostCenter  = "IT"
  }
}

data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"]

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}



resource "aws_instance" "web" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = "t3.micro"
  subnet_id              = module.vpc.public_subnets[0]
  vpc_security_group_ids = [aws_security_group.web.id]
  user_data              = <<-EOF
  #!/bin/bash
  echo "Hello, World!" > /var/www/html/index.html
  EOF

  tags = merge(local.common_tags, {
    Name = "web-server"
    Role = "web"
  })
}

resource "aws_security_group" "web" {
  name        = "web-sg"
  description = "Security group for web server"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "web-security-group"
  }
}
module "vpc" {
  source             = "./modules/vpc"
  vpc_cidr           = "10.0.0.0/16"
  environment        = "production"
  enable_nat_gateway = true
}

output "instance_id" {
  value       = aws_instance.web.id
  description = "The ID of the web server"
}

    #!/bin/bash






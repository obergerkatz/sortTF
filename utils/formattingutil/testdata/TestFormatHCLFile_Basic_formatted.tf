# This is a poorly formatted Terraform file
terraform {
  required_version = ">= 1.0"
  backend "s3" {
    bucket = "terraform-state-bucket"
    key    = "prod/terraform.tfstate"
  }
}

provider "aws" {
  region = var.region
}

variable "region" {
  type    = string
  default = "us-west-2"
}

locals {
  common_tags = {
    Owner       = "DevOps Team"
    Environment = "production"
  }
}

resource "aws_instance" "web" {
  ami           = "ami-123456"
  instance_type = "t3.micro"

  tags = {
    Name = "web-server"
  }
}








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



data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }
}



resource "aws_instance" "app" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = "t3.small"

  tags = merge(local.common_tags, {
    Name = "app-server"
  })
}
resource "aws_instance" "web" {
  ami           = "ami-123456"
  instance_type = "t3.micro"

  tags = {
    Name = "web-server"
  }
}
resource "aws_s3_bucket" "data" {
  bucket = "my-data-bucket"

  tags = {
    Environment = "production"
    Project     = "data-processing"
  }
}











module "vpc" {
  source = "./modules/vpc"

  vpc_cidr = "10.0.0.0/16"
}
output "bucket_name" {
  value = aws_s3_bucket.data.bucket
}












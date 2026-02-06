terraform {
  required_version = ">= 1.0"
}

provider "aws" {
  region = "us-west-2"
}

variable "name" {
  type = string
}

locals {
  tags = {}
}

data "aws_ami" "ubuntu" {
  most_recent = true
}

resource "aws_instance" "web" {
  ami = "ami-123"
}

module "vpc" {
  source = "./vpc"
}

output "id" {
  value = aws_instance.web.id
}

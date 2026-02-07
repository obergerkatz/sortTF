resource "aws_instance" "web" {
  instance_type = "t3.micro"
  ami           = "ami-123456"
  tags = {
    Name = "web-server"
  }
}

provider "aws" {
  region = "us-west-2"
}

variable "environment" {
  type = string
}

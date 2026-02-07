provider "aws" {
  region = "us-west-2"
}

variable "environment" {
  type = string
}

resource "aws_instance" "web" {
  ami           = "ami-123456"
  instance_type = "t3.micro"
  tags = {
    Name = "web-server"
  }
}

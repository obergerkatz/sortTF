variable "env" {
  default = "prod"
}

resource "aws_instance" "web" {
  ami = "ami-123"
  tags = {
    Name        = "web-${var.env}"
    Environment = var.env
  }
}

variable "enabled_true" {
  default = true
}

variable "enabled_false" {
  default = false
}

resource "aws_instance" "web" {
  ami                         = "ami-123"
  associate_public_ip_address = true
  monitoring                  = false
}

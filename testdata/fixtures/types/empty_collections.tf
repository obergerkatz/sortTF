variable "empty_list" {
  default = []
}

variable "empty_map" {
  default = {}
}

resource "aws_instance" "web" {
  ami             = "ami-123"
  tags            = {}
  security_groups = []
}

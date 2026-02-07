variable "nullable" {
  type    = string
  default = null
}

resource "aws_instance" "web" {
  ami                  = "ami-123"
  user_data            = null
  iam_instance_profile = null
}

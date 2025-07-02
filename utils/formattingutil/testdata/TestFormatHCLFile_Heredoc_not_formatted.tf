resource "aws_instance" "web" {
  ami       = "ami-123456"
  user_data = <<EOF
  #!/bin/bash
  echo Hello
  EOF
}
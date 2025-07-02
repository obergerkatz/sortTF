resource "aws_instance" "example" {
  ami = "ami-123456"

  provisioner "local-exec" {
    command = "echo hello"
  }

  lifecycle {
    create_before_destroy = true
  }
}
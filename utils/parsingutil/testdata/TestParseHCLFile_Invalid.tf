resource "aws_instance" "example" {
  ami = "ami-123456"
  invalid_attribute
  
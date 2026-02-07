resource "aws_instance" "web" {
  count         = var.enabled ? 1 : 0
  ami           = "ami-123"
  instance_type = var.instance_type != "" ? var.instance_type : "t2.micro"
}

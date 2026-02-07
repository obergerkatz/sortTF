resource "aws_instance" "web" {
  for_each      = var.instances
  ami           = "ami-123"
  instance_type = each.value.type
  tags = {
    Name = each.key
  }
}

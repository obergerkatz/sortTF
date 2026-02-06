resource "aws_instance" "web" {
  count         = var.instance_count
  ami           = "ami-123"
  instance_type = "t2.micro"
  tags = {
    Name = "web-${count.index}"
  }
}

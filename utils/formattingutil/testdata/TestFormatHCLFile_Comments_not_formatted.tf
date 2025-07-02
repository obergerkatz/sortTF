# This is a comment
resource "aws_instance" "web" {
  # inline comment
  ami           = "ami-123456"
  instance_type = "t3.micro"
  # another comment
  tags = {
    Name = "web-server"
    # end of block
  }
}
# final comment 
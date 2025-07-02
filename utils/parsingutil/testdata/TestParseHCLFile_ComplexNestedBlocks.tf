resource "aws_instance" "example" {
  ami = "ami-123456"

  dynamic "ebs_block_device" {
    for_each = var.ebs_block_devices
    content {
      device_name = ebs_block_device.value.device_name
      volume_size = ebs_block_device.value.volume_size

      dynamic "encryption" {
        for_each = ebs_block_device.value.encrypted ? [1] : []
        content {
          enabled = true
        }
      }
    }
  }
}

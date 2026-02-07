locals {
  tags = merge(
    {
      Environment = var.env
    },
    var.common_tags
  )

  cidr_blocks = [
    for i in range(3) :
    "10.0.${i}.0/24"
  ]

  uppercase_name = upper(var.name)
  joined_names   = join(",", var.names)
}

locals {
  user_data = <<USERDATA
#!/bin/bash
echo "Hello World"
USERDATA
}

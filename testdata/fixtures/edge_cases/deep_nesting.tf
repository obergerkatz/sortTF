# Deep nesting edge case - tests sorting deeply nested blocks
resource "null_resource" "deeply_nested" {
  provisioner "local-exec" {
    command = "echo 'level 1'"

    connection {
      type = "ssh"

      timeout {
        create = "5m"

        retry {
          max_attempts = 3

          backoff {
            initial_delay = "1s"

            exponential {
              multiplier = 2

              limits {
                max_delay = "30s"

                circuit_breaker {
                  threshold = 5

                  recovery {
                    timeout = "1m"

                    monitoring {
                      enabled = true

                      alerts {
                        email = "ops@example.com"
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}

# Test that attributes at each level are sorted
resource "example" "nested_attrs" {
  zulu = "last"
  alpha = "first"

  nested {
    yankee = "last"
    bravo = "first"

    deeper {
      zebra = "last"
      charlie = "first"

      deepest {
        xyz = "last"
        abc = "first"
      }
    }
  }
}

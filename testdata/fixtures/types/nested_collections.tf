variable "complex" {
  type = map(object({
    name   = string
    ports  = list(number)
    config = map(string)
  }))
  default = {
    app = {
      name = "web"
      ports = [80, 443]
      config = {
        env = "prod"
      }
    }
  }
}

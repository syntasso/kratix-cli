variable "name" {
  type        = string
  description = "Name for the resource"
}

variable "size" {
  type    = number
  default = 1
}

variable "tags" {
  type    = map(string)
  default = {}
}

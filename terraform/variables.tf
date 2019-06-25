variable "envoy_release" {
  description = "docker tag of envoy release"
  default     = "v1.10.0"
}

variable "release" {
  description = "roxprox release"
}

variable "acme_contact" {
  default     = "email@invalid.inv"
  description = "email address to be used for ACME - Let's encrypt will use this to notify you of expiring domains"
}

variable "control_plane_count" {
  description = "number of control plane instances to run"
  default     = 1
}

variable "envoy_proxy_count" {
  description = "number of envoy proxies to run"
  default     = 1
}

variable "subnets" {
  type        = list(string)
  description = "subnets to use"
}

variable "s3_bucket" {
  description = "name of s3 bucket to use"
}

variable "envoy_autocert_loglevel" {
  description = "log level"
  default     = "info"
}

variable "loadbalancer" {
  description = "loadbalancer type to use"
  default     = "nlb"
}

variable "loadbalancer_alb_cert" {
  description = "loadbalancer alb certificate to use"
  default     = ""
}

variable "tls_listener" {
  description = "run a service for a tls (https) listener (true/false)"
  type        = bool
}

variable "envoy_release" {
  description = "docker tag of envoy release"
  default     = "v1.10.0"
}

variable "release" {
  description = "roxprox release"
}

variable "acme_contact" {
  default     = ""
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

variable "envoy_proxy_cpu" {
  description = "fargate task cpu"
  default     = 256
}
variable "envoy_proxy_memory" {
  description = "fargate task memory"
  default     = 512
}

variable "envoy_proxy_appmesh_cpu" {
  description = "fargate task cpu when appmesh is enabled"
  default     = 512
}
variable "envoy_proxy_appmesh_memory" {
  description = "fargate task memory when appmesh is enabled"
  default     = 1024
}


variable "subnets" {
  type        = list(string)
  description = "subnets to use"
}

variable "lb_subnets" {
  type        = list(string)
  description = "loadbalancer subnets to use"
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

variable "loadbalancer_healthcheck_matcher" {
  description = "loadbalancer healthcheck matcher to use"
  default     = "200,404,301,302"
}

variable "loadbalancer_healthcheck_path" {
  description = "loadbalancer healthcheck path to use"
  default     = "/"
}

variable "loadbalancer_alb_cert" {
  description = "loadbalancer alb certificate to use"
  default     = ""
}
variable "loadbalancer_ssl_policy" {
  description = "ssl policy for the https listener to use"
  default     = "ELBSecurityPolicy-2016-08"
}
variable "loadbalancer_https_forwarding" {
  description = "if true, redirect all http traffic to https"
  default     = false
}

variable "tls_listener" {
  description = "run a service for a tls (https) listener (true/false)"
  type        = bool
}

variable "management_access_sg" {
  description = "allow access to the management interface"
  type        = list
  default     = []
}

variable "enable_appmesh" {
  description = "enable app mesh"
  type        = bool
  default     = false
}

variable "appmesh_name" {
  description = "name of the app mesh"
  default     = ""
}

variable "appmesh_envoy_release" {
  description = "tag of appmesh envoy release"
  default     = "v1.11.1.1-prod"
}

variable "appmesh_backends" {
  description = "list of backends to be configured in the appmesh virtual node"
  type        = list
  default     = []
}

variable "extra_containers" {
  description = "add extra containers to task definition"
  default     = ""
}

variable "extra_dependency" {
  description = "add extra dependencies to task definition"
  default     = ""
}

variable "extra_task_execution_policy" {
  description = "extra task execution policy for roxprox"
  default     = ""
}

variable "extra_task_role_policy" {
  description = "extra task role policy for roxprox"
  default     = ""
}

variable "enable_datadog" {
  description = "flag to enable datadog integration"
  default     = false
}
variable "datadog_api_key" {
  description = "datadog api key"
  default     = ""
}
variable "datadog_stats_url" {
  description = "datadog stats url"
  default     = ""
}
variable "datadog_image" {
  description = "datadog agent image"
  default     = "datadog/agent"
}
variable "datadog_image_version" {
  description = "datadog agent image version"
  default     = "latest"
}
variable "datadog_count" {
  description = "datadog service count"
  default     = "2"
}
variable "datadog_extra_task_execution_policy" {
  description = "datadog extra task execution policy"
  default     = ""
}
variable "datadog_env" {
  description = "datadog APM default enviroment"
  default     = "none"
}
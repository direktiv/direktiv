variable "IMAGE" {
  type = string
}
variable "UI_BACKEND" {
  type = string
}


job "DEPLOYMENT_NAME" {
  datacenters = ["dev"]

  group "direktiv-ui" {
    network {
      port "http" {
        to = 1644
      }
    }

    task "direktiv-ui" {
      driver = "docker"

      config {
        image      = var.IMAGE
        force_pull = true
        ports = ["http"]
      }

      env {
        UI_PORT    = 1644
        UI_BACKEND = var.UI_BACKEND
      }

      service {
        name     = "direktiv-ui-DEPLOYMENT_NAME"
        port     = "http"
        provider = "nomad"

        check {
          type     = "http"
          protocol = "http"
          path     = "/"
          interval = "2s"
          timeout  = "2s"
        }
      }
    }
  }
}

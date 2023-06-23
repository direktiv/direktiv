variable "IMAGE" {
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

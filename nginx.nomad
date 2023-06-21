job "deployment-nginx" {
  datacenters = ["dev"]

  group "nginx" {
    network {
      port "https" {
        static = 443
      }
    }

    task "nginx" {
      driver = "docker"

      env {
      }

      config {
        image   = "nginx"
        volumes = ["local:/etc/nginx/conf.d", "/etc/nomad/certificates:/etc/nomad/certificates"]
        ports   = ["https"]
      }

      service {
        name     = "nginx"
        port     = "https"
        provider = "nomad"

        check {
          type     = "tcp"
          interval = "10s"
          timeout  = "2s"
        }
      }

      template {
        data = <<EOF


############################################################# Upstreams

{{ range nomadServices }}
{{ if .Name | contains "direktiv-ui" }}
{{$name := .Name -}}
{{ range nomadService $name }}

upstream {{ .Name }} {
  server {{ .Address }}:{{ .Port }};
}

{{ end }}
{{ end }}
{{ end }}

# resolver 8.8.8.8;
# error_log logs/error.log debug;


############################################################# Default Servers
  server {
      listen 80 default_server;
      listen [::]:80 default_server;
      server_name _;
      deny all;
      return 444;
  }
  server {
      listen 443 ssl;
      listen [::]:443 ssl;
      server_name _;
      ssl_certificate /etc/nomad/certificates/wild-cert.crt;
      ssl_certificate_key /etc/nomad/certificates/wild-cert.key;
      deny all;
      return 444;
  }
############################################################# Proxy Server
  server {
      server_name "~^([a-zA-Z0-9-_]*).direktiv.dev";
      set $deployment_name $1;

      listen 443 ssl;
      listen  [::]:443 ssl;

      ssl_certificate /etc/nomad/certificates/wild-cert.crt;
      ssl_certificate_key /etc/nomad/certificates/wild-cert.key;

      location / {
        proxy_pass http://${deployment_name};
        proxy_set_header Host $host;
        proxy_set_header   X-Forwarded-For    $proxy_add_x_forwarded_for;
        proxy_set_header   X-Real-IP          $remote_addr;
      }
  }

  EOF

        destination   = "local/nginx.conf"
        change_mode   = "signal"
        change_signal = "SIGHUP"
      }
    }
  }
}

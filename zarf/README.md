# Zarf Air-Gapped Deployment Configuration

The primary `zarf.yaml` file defines the air-gapped deployment configuration for Direktiv. The YOLO (live) version's package configuration is stored in the `yolo` directory.

Component-specific configurations are located in the `components` folder. Each component's `zarf.yaml` serves as the base configuration, with `images/` and `yolo/` directories containing installation method-specific artifacts. The distinction between these configurations lies in the `image` section, which is tailored for air-gapped deployments.

**TEMPLATES**

The Direktiv component uses templates to update the version number during releases. The `zarf.yaml` file is generated using the Taskfile during package creation. Any Zarf changes should be made in the `zarf.template.yaml` files, not directly in the `zarf.yaml` files themselves. It replaces the `VERSION` tag and the Direktiv image in the air-gapped zarf configuration.


## Initialization for AirGap

Although Zarf can install K3S, we are assuming an existing installation of K3S. For Zarf to work, the cluster has to be initialized. For this process, it needs a `init` package which can be downloaded with `zarf tools download-init` and will be stored in `~/.zarf-cache/`. In an air-gapped environment, this file and the Zarf binary have to be put on the server. The following command initializes Zarf with a Git server.

`zarf init --components=git-server --confirm`

To get the credentials for the installed components, use `zarf tools get-creds` and to remove Zarf from the cluster, use `zarf destroy --confirm`.

***For YOLO (non-airgap) installations, there is no initialization of Zarf in the cluster needed.***

## Building Packages

This project uses `taskfile.dev` for building artifacts. One command builds an airgap package `zarf:create` and a second command `zarf:yolo-create` can build a YOLO package without dependencies. If no parameters are provided the default version will be `dev` and the image source for Direktiv will be set to the local repository. For production releases the version can be set with `VERSION` like the following command:

`task zarf:yolo-create VERSION=v0.9.1`

***The image and the Helm charts have to be released before building the package.***

## Deploying Packages

Packages can be deployed with one command `zarf package deploy zarf-package-direktiv-full-XXX`. Deployment parameters can be set with a `--set` argument but a safer way of providing deployment variables is using a `zarf-config.yaml`. 

TLS can be enabled by using `DIREKTIV_WITH_CERTIFICATE`. The Zarf installer uses `server.key` amd `server.crt` as server certificates. If `server.key` does not exist a self-signed certificate will be generated. If certificates are being used the value `DIREKTIV_HOST` has to be set to the DNS name of the cluster. 

### Installation configuration

Deployments can be configured with a `zarf-config.yaml` file in the installation directory. The components to be installed and installation values can be configured. The following is an example how to install a development version with Zarf to the cluster: 

```
log_level : 'info'

package:
  deploy:
    components: 'direktiv,postgres,linkerd'
    set:
      direktiv_request_timeout: 28800
      direktiv_ingress_hostport: "true"
      direktiv_ingress_service_type: ClusterIP
      direktiv_tag: dev
      direktiv_registry: localhost:5001      
      direktiv_function_sizes: |
        limits:
          memory:
            small: 256
            medium: 512
            large: 2048
          cpu:
            small: 300m
            medium: 450m
            large: 2000m
          disk:
            small: 128
            medium: 256
            large: 1024
      direktiv_image: direktiv
```

### Developing Packages

To develop and change the Zarf package there are three commands to reset the KIND cluster and deploy a package:

- `cluster:one-node`
- `zarf-create` or `zarf-yolo-create`
- `zarf:deploy` or `zarf:yolo-deploy`



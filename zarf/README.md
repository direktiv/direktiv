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


certificates / direktiv-host

### Installation configuration

### Developing Packages

To develop and change the Zarf package there are two commands to reset the KIND cluster and deploy a package:

- `cluster:one-node`
- `zarf:deploy` or `zarf:yolo-deploy`



# INIT

## WITH k3S

`sudo zarf init --components=git-server,k3s --confirm`

To delete k3s: 

`sudo zarf destroy --confirm`

## WITHOUT k3S

`zarf init --components=git-server --confirm`

## COPY KUBECONFIG (INTERNAL K3S)

```sh
sudo cp /root/.kube/config /home/{MYUSER}/.kube
sudo chmod 644 /home/{MYUSER}/.kube/config
export KUBECONFIG=/home/{MYUSER}/.kube/config
```

# COMMANDS

## FIND IMAGES FOR COMPONENT

`zarf dev find-images .`

## GET SERVICES AND CREDENTIALS FOR ZARF

`zarf tools get-creds`

## CREATE AND DEPLOY COMPONENTS

`zarf package create .`

`zarf package deploy`

## ADDITIONAL CONFIG

The individual configuration of the components can be found in the README files of in the `components` folder. The configuration items there can be set via the command line `--set` argument or in a `zarf-config.yaml` file. 

The configuration values for the `init` package can be found here:

https://docs.zarf.dev/tutorials/package_create_init.html

With those variables the e.g. stroage or CPU values can be changed for the Git server or the image repository.
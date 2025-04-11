# INIT

## WITH k3S

sudo zarf init --components=git-server,k3s --confirm

## WITHOUT k3S

zarf init --components=git-server --confirm

## COPY KUBECONFIG

```sh
sudo cp /root/.kube/config /home/{MYUSER}/.kube
sudo chmod 644 /home/{MYUSER}/.kube/config
export KUBECONFIG=/home/{MYUSER}/.kube/config
```

## ADDITIONAL CONFIG

https://docs.zarf.dev/tutorials/package_create_init.html


# COMMANDS

## FIND IMAGES FOR COMPONENT

`zarf dev find-images .`

## GET SERVICES AND CREDENTIALS FOR ZARF

`zarf tools get-creds`

## CREATE AND DEPLOY COMPONENTS

`zarf package create .`
`zarf package deploy`




# INSTALL ON EXISTING k3S


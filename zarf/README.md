# INIT

## WITH k3S

sudo zarf init --components=git-server,k3s --confirm

sudo cp /root/.kube/config /home/{MYUSER}/.kube
sudo chmod 644 /home/{MYUSER}/.kube/config
export KUBECONFIG=/home/{MYUSER}/.kube/config

## WITHOUT  k3S

sudo zarf init --components=git-server --confirm

## ADDITIONAL CONFIG

https://docs.zarf.dev/tutorials/package_create_init.html


# COMMANDS

zarf dev find-images .


zarf package create .
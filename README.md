# direktiv

<br />
<p align="center">
  <a href="https://github.com/vorteil/direktiv">
    <img src="assets/images/direktiv-logo.png" alt="vorteil">
  </a>
    <h5 align="center">event-based serverless container workflows</h5>
</p>
<hr/>


[![Build](https://github.com/vorteil/direktiv/actions/workflows/build.yml/badge.svg)](https://github.com/vorteil/direktiv/actions/workflows/build.yml) <a href="https://codeclimate.com/github/vorteil/direktiv/maintainability"><img src="https://api.codeclimate.com/v1/badges/39969b6bb893928434ae/maintainability" /></a> [![Go Report Card](https://goreportcard.com/badge/github.com/vorteil/direktiv)](https://goreportcard.com/report/github.com/vorteil/direktiv) [![Discord](https://img.shields.io/badge/chat-on%20discord-6A7EC2)](https://discord.gg/VjF6wn4)

> 
>
> **Check out our online demo: [wf.direktiv.io](https://wf.direktiv.io)**
> 
>  

## What is Direktiv?

Direktiv is a specification for a serverless computing workflow language that aims to be simple and powerful above all else.

Direktiv defines a selection of intentionally primitive states, which can be strung together to create workflows as simple or complex as the author requires. The powerful `jq` JSON processor allows authors to implement sophisticated control flow logic, and when combined with the ability to run containers as part of Direktiv workflows just about any logic can be implemented. 

Workflows can be triggered by CloudEvents for event-based solutions, can use cron scheduling to handle periodic tasks, and can be scripted using the APIs for everything else.

## Why use Direktiv?

Direktiv was created to address 4 problems faced with workflow engines in general:

- *Cloud agnostic*: we wanted Direktiv to run on any platform or cloud, support any code or capability and NOT be dependent on the cloud provider's services for running the workflow or executing the actions (but obviously support it all)
- *Simplicity*: the configuration of the workflow components should be simple more than anything else. Using only YAML and `jq` you should be able to express all workflow states, transitions, evaluations and actions needed to complete the workflow
- *Reusable*: if you're going to the effort and trouble of pushing all your microservices, code or application components into a container platform why not have the ability to reuse and standardise this code across all of your workflows. We wanted to ensure that your code always remains reusable and portable and not tied into a specific vendor format or requirement (or vendor specific language) - so we've modelled Direktiv's specification after the [CNCF Serverless Workflow Specification](https://github.com/serverlessworkflow/specification) with the ultimate goal to implement it fully
- *Multi-tenanted and secure*: we want to use Direktiv in a multi-tenant service provider space, which means all workflow executions have to be isolated, data access secured and isolated and all workflows and actions are truly ephemeral (or serverless).

## Direktiv internals?
This repository contains a reference implementation that runs Docker containers as isolated virtual machines on [Firecracker](https://github.com/firecracker-microvm/firecracker) using [Vorteil.io](github.com/vorteil/vorteil).

<p align="center">
  <img src="assets/images/direktiv-overview-solid.png" alt="direktiv">
</p>



## Quickstart

### Starting the Server

Getting a local playground environment can be easily done with either [Vorteil.io](github.com/vorteil/vorteil) or Docker:

****

***Using Docker:***

`docker run --net=host --privileged vorteil/direktiv`. 

*Note: *

- *You may need to run this command as an administrator.*

- *In a public cloud instance, nested virualization is needed to support the firecracker micro-VMs. Each public cloud provider has different configuration settings which need to be applied to enable nested virtualization. Examples are shown below for each public cloud provider:*
  - [Google Cloud Platform](https://cloud.google.com/compute/docs/instances/enable-nested-virtualization-vm-instances)
  - Amazon Web Services (only supported on bare metal instances)
  - [Microsoft Azure](https://docs.microsoft.com/en-us/azure/virtual-machines/windows/nested-virtualization)
  - Alibaba (only supported on bare metal instances)
  - [Oracle Cloud](https://blogs.oracle.com/cloud-infrastructure/nested-kvm-virtualization-on-oracle-iaas)
  - [VMware](https://communities.vmware.com/t5/Nested-Virtualization-Documents/Running-Nested-VMs/ta-p/2781466)



***Using Vorteil:***

With Vorteil installed (full instructions [here](https://github.com/vorteil/vorteil)):

 1. download `direktiv.vorteil` from the [releases page](https://github.com/vorteil/direktiv/releases), 
 2. run `vorteil run direktiv.vorteil` from within your downloads folder.



***Testing Direktiv***:

Download the `direkcli` command-line tool from the [releases page](https://github.com/vorteil/direktiv/releases)  and create your first namespace by running:

`direkcli namespaces create demo`

```bash
$ direkcli namespaces create demo
Created namespace: demo
$ direkcli namespaces list
+------+
| NAME |
+------+
| demo |
+------+
```



### Workflow specification

The below example is the minimal configuration needed for a workflow, following the [workflow language specification](https://docs.direktiv.io/docs/specification.html): 

```yaml
id: helloworld
states:
- id: hello
  type: noop
  transform: '{ msg: ("Hello, " + .name + "!") }'
```



### Creating and Running a Workflow

The following script does everything required to run the first workflow. This includes creating a namespace & workflow and running the workflow the first time.  

```bash
$ direkcli namespaces create demo
Created namespace: demo
$ cat > helloworld.yml <<- EOF
id: helloworld
states:
- id: hello
  type: noop
  transform: '{ msg: ("Hello, " + .name + "!") }'
EOF
$ direkcli workflows create demo helloworld.yml
Created workflow 'helloworld'
$ cat > input.json <<- EOF
{
  "name": "Alan"
}
EOF
$ direkcli workflows execute demo helloworld --input=input.json
Successfully invoked, Instance ID: demo/helloworld/aqMeFX <---CHANGE_THIS_TO_YOUR_VALUE
$ direkcli instances get demo/helloworld/aqMeFX
ID: demo/helloworld/aqMeFX
Input: {
  "name": "Alan"
}
Output: {"msg":"Hello, Alan!"}
```



### Code of Conduct

We have adopted the [Contributor Covenant](https://github.com/vorteil/.github/blob/master/CODE_OF_CONDUCT.md) code of conduct.

### Contributing

Any feedback and contributions are welcome. Read our [contributing guidelines](https://github.com/vorteil/.github/blob/master/CONTRIBUTING.md) for details.

## License

Distributed under the Apache 2.0 License. See `LICENSE` for more information.

## See Also

* The [direktiv.io](https://direktiv.io/) website.
* The [vorteil.io](https://github.com/vorteil/vorteil/) repository.
* The Direktiv [documentation](https://docs.direktiv.io/).
* The [Direktiv Beta UI](http://wf.direktiv.io/).
* The [Godoc](https://godoc.org/github.com/vorteil/direktiv) library documentation.

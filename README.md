# direktiv

<br />

<p align="center">
  <a href="https://github.com/direktiv/direktiv">
    <img src="assets/images/direktiv-logo.png" alt="direktiv">
  </a>
    <h5 align="center">event-driven serverless orchestration</h5>
</p>
<hr/>

[![Build](https://github.com/direktiv/direktiv/actions/workflows/build.yml/badge.svg)](https://github.com/direktiv/direktiv/actions/workflows/build.yml) <a href="https://codeclimate.com/github/direktiv/direktiv/maintainability"><img src="https://api.codeclimate.com/v1/badges/39969b6bb893928434ae/maintainability" /></a> [![Go Report Card](https://goreportcard.com/badge/github.com/direktiv/direktiv)](https://goreportcard.com/report/github.com/direktiv/direktiv) [![Discord](https://img.shields.io/badge/chat-on%20discord-6A7EC2)](https://discord.gg/VjF6wn4)


# What is direktiv?

Direktiv is an event-driven container orchestration engine, running on Kubernetes and Knative. The following key concepts:

- direktiv runs containers as part of workflows from any compliant container registry, passing JSON structured data between workflow states.
- JSON structured data is passed to the containers using HTTP protocol on port 8080.
- direktiv uses a [primitive state declaration specification](https://docs.direktiv.io/latest/specification) to describe the flow of the orchestration in YAML, or users can build the workflow using the workflow builder UI.
- direktiv uses `jq` JSON processor to implement sophisticated control flow logic and data manipulation through states.
- Workflows can be event-based triggers ([Knative Eventing](https://knative.dev/docs/eventing/) & [CloudEvents](https://cloudevents.io/)), cron scheduling to handle periodic tasks, or can be scripted using the APIs.
- Integrated into [Prometheus](https://prometheus.io/) (metrics), [Fluent Bit](https://fluentbit.io/) (logging) & [OpenTelemetry](https://opentelemetry.io/) (instrumentation & tracing).

Additional resources to get started:

- Pre-built plugins are available from [https://github.com/direktiv/direktiv-apps](https://github.com/direktiv/direktiv-apps) - we're working hard to add more every day!
- Examples for integration your own containers [https://github.com/direktiv/direktiv-apps/tree/main/examples](https://github.com/direktiv/direktiv-apps/tree/main/examples) with an explanation [here](https://docs.direktiv.io/latest/getting_started/making-functions/).

<table>
  <tr>
    <th>Dashboard</th>
    <th>Flow Builder</th>
  </tr>
  <tr>
    <td><img src="assets/images/direktiv-ui.png" alt="direktiv ui"></td>
    <td><img src="assets/images/workflow-builder.png" alt="direktiv ui"></td>
  </tr>
  <tr>
    <th>YAML definition</th>
    <th>OpenTelemetry Integration</th>
  </tr>
  <tr>
    <td><img src="assets/images/yaml.png" alt="direktiv ui"></td>
    <td><img src="assets/images/grafana-tempo.png" alt="direktiv ui"></td>
  </tr>  
</table>


# Why use direktiv?

- *Cloud agnostic*: direktiv runs on any platform, supports any code and is not dependent on cloud provider's services for running workflows or executing actions
- *Simplicity*: the configuration of the workflow components should be simple more than anything else. Using only YAML and `jq` you should be able to express all workflow states, transitions, evaluations, and actions needed to complete the workflow
- *Reusable*: if you're going to the effort and trouble of pushing all your microservices, code or application components into a container platform why not have the ability to reuse and standardise this code across all your workflows? We wanted to ensure that code always remains reusable and portable without the need for SDKs (or vendor specific language).

# Quickstart

## Running a local direktiv instance (docker)

Getting a local playground environment can be easily done with Docker. The following command starts a docker container with kubernetes. *On startup it can take a few minutes to download all images.* When the installation is done all pods should show "Running" or "Completed".

```sh
docker run --privileged -p 8080:80 -ti direktiv/direktiv-kube
```

***Testing Installation:***

Browse the UI: http://localhost:8080

... or ...

verify direktiv is online manually from the command-line using `cURL`:

```sh
$ curl -vv -X PUT "http://localhost:8080/api/namespaces/demo"
{
  "namespace": {
    "createdAt": "2021-10-06T00:03:22.444884147Z",
    "updatedAt": "2021-10-06T00:03:22.444884447Z",
    "name": "demo",
    "oid": ""
  }
}
```

## Kubernetes Install

For full instructions on how to install direktiv on a Kubernetes environment go to the [installation pages](https://docs.direktiv.io/latest/installation/)


## Creating your first workflow

The following script does everything required to run the first workflow. This includes creating a namespace & workflow and running the workflow the first time.  

```bash
$ curl -X PUT "http://localhost:8080/api/namespaces/demo"
{
  "namespace": {
    "createdAt": "2021-10-06T00:03:22.444884147Z",
    "updatedAt": "2021-10-06T00:03:22.444884447Z",
    "name": "demo",
    "oid": ""
  }
}
$ cat > helloworld.yml <<- EOF
states:
- id: hello
  type: noop
  transform:
    msg: "Hello, jq(.name)!"
EOF
$ curl -vv -X PUT --data-binary "@helloworld.yml" "http://localhost:8080/api/namespaces/demo/tree/helloworld?op=create-workflow"
{
  "namespace": "demo",
  "node": {...},
  "revision": {...}
}
$ cat > input.json <<- EOF
{
  "name": "Alan"
}
EOF
$ curl -vv -X POST --data-binary "@input.json" "http://localhost:8080/api/namespaces/demo/tree/helloworld?op=wait"
{"msg":"Hello, Alan!"}
```

## Running a container workflow

The next example uses the direktiv/request container in [https://hub.docker.com/r/direktiv/request](https://hub.docker.com/r/direktiv/request). The container starts a HTTP listener on port 8080 and accepts as input a JSON object containing all the parameters for a HTTP(S) request. It returns the result to the workflow on the same listener. This is the template for how all containers run as part of workflow execution.

```bash
$ cat > bitcoin.yaml <<- EOF
functions:
  - type: reusable
    id: get-request
    image: direktiv/request:latest
states:
  - id: get-bitcoin
    type: action
    log: jq(.)
    action:
      function: get-request
      input:
        method: "GET"
        url: "https://blockchain.info/ticker"
      retries:
        max_attempts: 3
        delay: PT30S
        multiplier: 2.0
        codes: [".*"]
    transform: "jq({ value: .return.body.USD.last })"
    transition: print-bitcoin
  - id: print-bitcoin
    type: noop
    log: "BTC value: jq(.value)"
EOF
$ curl -vv -X PUT --data-binary "@bitcoin.yaml" "http://localhost:8080/api/namespaces/demo/tree/get-bitcoin?op=create-workflow"
{
  "namespace":  "demo",
  "node":  {... },
  "revision":  {...}
}
$ curl -X POST  "http://localhost:8080/api/namespaces/demo/tree/get-bitcoin?op=wait"
{
  "value":62988.71
}
```

The UI displays the log output and state of the workflow from start to completion.

<p align="center">
  <img src="assets/images/bitcoin_ui.png" alt="direktiv ui">
</p>

# Documentation

- [Getting Started](https://docs.direktiv.io/latest/getting_started/helloworld/)
- [Workflow Specification](https://docs.direktiv.io/latest/specification/)
- [Examples](https://docs.direktiv.io/latest/examples/greeting/)

# Talk to us!

- [Open Source Support Channel on Slack](https://direktivio.slack.com/archives/C02JQUH1A01)
- [Open Source Support Channel on Discord](https://discord.gg/VjF6wn4)


# Code of Conduct

We have adopted the [Contributor Covenant](https://github.com/direktiv/.github/blob/master/CODE_OF_CONDUCT.md) code of conduct.

# Contributing

Any feedback and contributions are welcome. Read our [contributing guidelines](https://github.com/direktiv/.github/blob/master/CONTRIBUTING.md) for details.

# License

Distributed under the Apache 2.0 License. See `LICENSE` for more information.

# See Also

* The [direktiv.io](https://direktiv.io/) website.
* The direktiv [documentation](https://docs.direktiv.io/).
* The direktiv [blog](https://blog.direktiv.io/).
* The [Godoc](https://godoc.org/github.com/direktiv/direktiv) library documentation.

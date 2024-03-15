<p align="center">
  <img src="assets/images/direktiv-logo-50.png" alt="direktiv">
</p>

<br>

<div align="center">

[![License](https://img.shields.io/badge/License-Apache--2.0-blue)](#license)
[![Go Report Card](https://goreportcard.com/badge/github.com/direktiv/direktiv)](https://goreportcard.com/report/github.com/direktiv/direktiv) 
[![GitHub release](https://img.shields.io/github/release/direktiv/direktiv.svg)](https://github.com/direktiv/direktiv/releases/)
[![GitHub stars](https://badgen.net/github/stars/direktiv/direktiv)](https://github.com/direktiv/direktiv/stargazers/)
[![GitHub contributors](https://img.shields.io/github/contributors/direktiv/direktiv.svg)](https://github.com/direktiv/badges/graphs/contributors/)
[![Maintenance](https://img.shields.io/badge/Maintained%3F-yes-green.svg)](https://github.com/direktiv/direktiv/graphs/commit-activity)
[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](https://go.dev/)
[![Slack](https://img.shields.io/badge/Slack-Join%20Direktiv-4a154b?style=flat&logo=slack)](https://join.slack.com/t/direktiv-io/shared_invite/zt-zf7gmfaa-rYxxBiB9RpuRGMuIasNO~g)

</div>


<h1 align="center">Event-Driven Serverless Orchestration, Integration and Automation</h1>
<div align="center">
Run Workflows and Create Services in Seconds
</div>
</br>

<p align="center">
    <a href="https://docs.direktiv.io">Documentation</a>
    ·
    <a href="https://blog.direktiv.io">Blog</a>
    ·
    <a href="https://github.com/direktiv/direktiv/issues/new?assignees=&labels=&projects=&template=bug_report.md&title=">Report Bug</a>
    ·
    <a href="https://github.com/direktiv/direktiv/issues/new?assignees=&labels=&projects=&template=feature_request.md&title=">Request Feature</a>
</p>


<div align="center"><img src="assets/images/direktiv.gif" ></div>

## Features & Standards

* __YAML__: Define flows and subflows with simple YAML including if/else switches, error handling, retries, validations and more.
* __Serverless__: Call multiple serverless functions from a flow and merge and modify the responses to a single consumable function call.
* __Event-Based__: Catch events within the system or from external sources like AWS or Github and execute flows based on that event.
* __JSON Inputs & States__: Use JSON as input for flows and respond with JSON to the caller. JSON will be saved between the states fo a flow.
* __API Gateway__: Includes an API gateway to expose flows as services for third-party consumers including authentication.
* __CloudEvents__: Supports CNCF's [CloudEvents](https://cloudevents.io/) natively.
* __GitOps Approach__: All configurations, services and flows can be synced from Git. Git becomes the single source of truth.
* __Observability__: Integrated into Prometheus (metrics), Fluent Bit (logging) & OpenTelemetry (instrumentation & tracing).
* __Periodic Tasks__: Call flows periodically via cron jobs for repeating tasks. 
* __Scalable__: Direktiv scales on mulitple levels with Kubernetes scaling and [Knative's](https://knative.dev/docs/) scaling features.
* __Easily Extendable__: Add custom functions with simple Docker containers.

## Quick Start

Direktiv provides a Docker container with all required components pre-installed (Linux only). The initial startup can take a couple of minutes to download all required images.

```sh
docker run --privileged -p 8080:80 -ti direktiv/direktiv-kube
```

> If the upper limit for inotify instances is too low the pods might be stuck in *pending*. Increase that limit if necessary with the command `sudo sysctl fs.inotify.max_user_instances=4096`

If you are not using Linux please follow the [installation instructions](https://docs.direktiv.io/installation/kubernetes/) on the documentation page. 



## About Direktiv

Direktiv is an event-driven workflow engine made for **orchestration**, **integration**, and **automation**. In it's core it is a state machine which uses containers as functions within workflows and passes JSON structured data between states. It offers key features like retries, error handling, and conditional logic. The flow's state, stored as JSON, allows for dynamic transformations during execution using JQ or JavaScript.

<table align="center">
  <tr>
  <td>

```yaml
direktiv_api: workflow/v1
states:
- id: start
  type: noop
  transform:
    result: Hello world!
  transition: second
- id: second
  type: noop
  log: second state
  transform:
    final: this value is from state one jq(.result)
```
  </td>
  <td>
    <img src="assets/images/flow.png" >
  </td>
  </tr>
</table>

Workflows can be triggered by events, start periodically via crons or can be started by a HTTP POST request where the data is the initial state fo the workflow. 

## Writing Workflows

Writing workflows is a very simple task. The basic idea is that there are multiple states where the workflow steps through. These states are of different types to provide switches, errors or other functionality. The most important type is the `action` state. This state calls a function which is basically a simple container.

Writing workflows is quite simple. Essentially, a workflow is progressing through different states with a JSOJ payload passed between those states. These states come in various types, providing switches, error management, and more. The most important type is the `action` state. The `action` state is basically a call to a container. Internally Direktiv spins up a serverless Knative function and passes data to that container. The response of that call becomes part of the JSON state of the flow.

```yaml
direktiv_api: workflow/v1

functions:
- id: request
  image: gcr.io/direktiv/functions/http-request:1.0
  type: knative-workflow

states:
- id: joke 
  type: action
  action:
    function: request
    input: 
      method: GET
      url: https://api.chucknorris.io/jokes/random?category=jq(.category // "dev")
  transform:
    joke: jq(.return[0].result.value)

```

The above example is a very simple one-step workflow. It uses the `http-request` function. This function accepts multiple input parameters e.g. the HTTP method or the URL. The section under `input` defines the data send to the function as JSON. In this case Direktiv would post the following payload to the container:

```json
{
  "method": "GET",
  "url": "https://api.chucknorris.io/jokes/random?category=dev"
}
```

In this example, there's a JQ statement, `jq(.category // "dev")`, which is getting evaluated at runtime. If the 'category' value is defined in the state, it's utilized in the URL; otherwise, `dev` is used as the default. For instance, initiating the workflow with a payload like `{ "category": "animal" }` would result in the URL: `https://api.chucknorris.io/jokes/random?category=animal`.

The state can be transformed after each state of the workflow. In this case the JQ command fetches the joke from the return values of the function and sets it as new state.

### Custom Functions


## See Also

* The [direktiv.io](https://direktiv.io/) website.
* The direktiv [documentation](https://docs.direktiv.io/).
* The direktiv [blog](https://blog.direktiv.io/).
* The [Godoc](https://godoc.org/github.com/direktiv/direktiv) library documentation.

<!-- # What is direktiv?

Direktiv is an event-driven container orchestration engine, running on Kubernetes and Knative. The following key concepts:

- direktiv runs containers as part of workflows from any compliant container registry, passing JSON structured data between workflow states.
- JSON structured data is passed to the containers using HTTP protocol on port 8080.
- direktiv uses a [primitive state declaration specification](https://docs.direktiv.io/spec/workflow-yaml/workflow/) to describe the flow of the orchestration in YAML, or users can build the workflow using the workflow builder UI.
- direktiv uses `jq` JSON processor to implement sophisticated control flow logic and data manipulation through states.
- Workflows can be event-based triggers ([Knative Eventing](https://knative.dev/docs/eventing/) & [CloudEvents](https://cloudevents.io/)), cron scheduling to handle periodic tasks, or can be scripted using the APIs.
- Integrated into [Prometheus](https://prometheus.io/) (metrics), [Fluent Bit](https://fluentbit.io/) (logging) & [OpenTelemetry](https://opentelemetry.io/) (instrumentation & tracing).

Additional resources to get started:

- Pre-built plugins are available from [https://github.com/direktiv/direktiv-apps](https://github.com/direktiv/direktiv-apps) - we're working hard to add more every day!
- Examples for integration your own containers [https://github.com/direktiv/direktiv-apps/tree/main/examples](https://github.com/direktiv/direktiv-apps/tree/main/examples) with an explanation [here](https://docs.direktiv.io/getting_started/).

<table>
  <tr>
    <th>Knative Services</th>
    <th>Orchestration Flow</th>
  </tr>
  <tr>
    <td><img src="assets/images/readme.png" alt="direktiv ui"></td>
    <td><img src="assets/images/readme1.png" alt="direktiv ui"></td>
  </tr>
  <tr>
    <th>YAML definition</th>
    <th>OpenTelemetry Integration</th>
  </tr>
  <tr>
    <td><img src="assets/images/readme2.png" alt="direktiv ui"></td>
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

> If the upper limit for inotify instances is too low the pods might be stuck in *pending*. Increase that limit if necessary: 

```console
sudo sysctl fs.inotify.max_user_instances=4096
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

For full instructions on how to install direktiv on a Kubernetes environment go to the [installation pages](https://docs.direktiv.io/installation/)


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
  <img src="assets/images/btc.png" alt="direktiv ui">
</p>

# Documentation

- [Getting Started](https://docs.direktiv.io/getting_started/)
- [Workflow Specification](https://docs.direktiv.io/spec/workflow-yaml/workflow/)
- [Examples](https://docs.direktiv.io/examples/aws/)

# Talk to us!

- [Open Source Support Channel on Slack](https://join.slack.com/t/direktiv-io/shared_invite/zt-zf7gmfaa-rYxxBiB9RpuRGMuIasNO~g)


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
* The [Godoc](https://godoc.org/github.com/direktiv/direktiv) library documentation. -->

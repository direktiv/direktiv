# Timeouts 

## TimeoutsDefinition

In addition to any timeouts applied on a state-by-state basis, every workflow has two global timeouts that begin ticking from the moment the workflow starts. This is where you can configure these timeouts differently to their defaults.

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `interrupt` | An ISO8601 duration string. Sets the time to wait before throwing a catchable `direktiv.cancels.timeout.soft` error. Consider this a soft timeout. | string | no |
| `kill` | An ISO8601 duration string. Sets the time to wait before throwing an uncatchable `direktiv.cancels.timeout.hard` error. This is a hard timeout. | string | no |

```yaml title="Workflow Timeout"
direktiv_api: workflow/v1
timeouts:
  interrupt: PT60M
  kill: PT30M
functions:
- type: knative-workflow
  id: request
  image: direktiv/request:latest
  size: small
states:
- id: getter
  type: action
  action:
    function: request
    input:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/todos/1"
```
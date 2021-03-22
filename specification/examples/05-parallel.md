# Parallel Execution and Wait Example

This example demonstrates the use of parallel subflows that must all complete before the `run` state will succeed.

A hypothetical scenario where this approach may be used could involve a CI/CD process for which 3 different binaries are built (one each on Windows, Linux, and Mac) before creating a new product release. The `run` workflow will wait until all three subflows have received an event before proceeding.


## Parallel Workflow YAML

```yaml
id: waiting
states:
- id: run
  type: parallel
  actions:
  - workflow: waitforwindows
  - workflow: waitforlinux
  - workflow: waitformac
  mode: and
```

## wait-for Workflow YAML

Replace `{OS}` with `windows`, `mac`, and `linux`, to create the 3 subflows referenced by the `run` state.

```yaml
id: wait-for-{OS}
states:
- id: waitForEvent
  type: consumeEvent
  event:
    type: gen-event-{OS}
```

## generateEvent Workflow YAML

Replace `{OS}` with `windows`, `mac` and `linux` to create workflows that will generate the events that the previous three subflows are waiting to receive.


```yaml
id: send-event-for-{OS}
states:
- id: sendEvent
  type: generateEvent
  event:
    type: gen-event-{OS}
    source: direktiv
```

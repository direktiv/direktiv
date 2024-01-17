# EventsXor State 

```yaml
direktiv_api: workflow/v1
states:
- id: a
  type: eventsXor
  timeout: PT15M
  events:
  - event:
      type: com.github.pull.create
      context:
        subject: '123'
    transition: received
    transform:
      hello: world
  - event:
      type: com.github.pull.delete
      context:
        subject: '123'
    transition: received

- id: received
  type: noop
```

## EventsXorStateDefinition 

To pause the workflow and wait until one of multiple CloudEvents events is received before proceeding, the `eventsXor` state might be used. Any event match received will cause this state to complete.

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StateDefinition](./states.md) is being used. In this case it must be set to `eventsXor`. | string | yes | 
| `id` | An identifier unique within the workflow to this one state. | string | yes |
| `log` | If defined, the workflow will generate a log when it commences this state. See [StateLogging](./logging.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `metadata` | If defined, updates the instance's metadata. See [InstanceMetadata](./metadata.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `catch` | Defines behaviour for handling of catchable errors.  | [[]ErrorCatchDefinition](errors.md#errorcatchdefinition) | no |
| `timeout` | An ISO8601 duration string. | string | no |
| `events` | Defines the criteria by which incoming CloudEvents events are evaluated to find a match. | [ConsumeEventDefinition](./consume-event.md#consumeeventdefinition) | yes |

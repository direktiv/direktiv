# EventsXor State 

```yaml
- id: a
  type: eventsXor
  timeout: PT15M
  events:
  - type: com.github.pull.create
	context:
	  subject: '123'
  - type: com.github.pull.delete
	context:
	  subject: '123'
```

## EventsXorStateDefinition 

To pause the workflow and wait until one of multiple CloudEvents events is received before proceeding, the `eventsXor` state might be used. Any event match received will cause this state to complete.

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StateDefinition](./states.md) is being used. In this case it must be set to `eventsXor`. | string | yes | 
| `id` | An identifier unique within the workflow to this one state. | string | yes |
| `log` | If defined, the workflow will generate a log when it commences this state. See [StateLogging](./logging.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `metadata` | If defined, updates the instance's metadata. See [InstanceMetadata](./metadata.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `transform` | If defined, modifies the instance's data upon completing the state logic. See [StateTransforms](../instance-data/transforms.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `transition` | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, this state terminates the workflow. | string | no |
| `catch` | Defines behaviour for handling of catchable errors.  | [[]ErrorCatchDefinition](./errors.md) | no |
| `timeout` | An ISO8601 duration string. | string | no |
| `events` | Defines the criteria by which incoming CloudEvents events are evaluated to find a match. | [ConsumeEventDefinition](./consume-event.md#ConsumeEventDefinition) | yes |

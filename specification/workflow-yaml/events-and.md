# EventsAnd State 

```yaml
- id: a
  type: eventsAnd
  timeout: PT15M
  events:
  - type: com.github.pull.create
    context:
      subject: '123'
  - type: com.github.pull.delete
    context:
      subject: '123'
```

## EventsAndStateDefinition 

To pause the workflow and wait until multiple CloudEvents events are received before proceeding, the `eventsAnd` is used. Every listed event must be received for the state to complete. If there are multiple events of the same type a index number will be added to the duplicate cloudevent types.

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StateDefinition](./states.md) is being used. In this case it must be set to `eventsAnd`. | string | yes | 
| `id` | An identifier unique within the workflow to this one state. | string | yes |
| `log` | If defined, the workflow will generate a log when it commences this state. See [StateLogging](./logging.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `metadata` | If defined, updates the instance's metadata. See [InstanceMetadata](./metadata.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `transform` | If defined, modifies the instance's data upon completing the state logic. See [StateTransforms](../instance-data/transforms.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `transition` | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, this state terminates the workflow. | string | no |
| `catch` | Defines behaviour for handling of catchable errors.  | [[]ErrorCatchDefinition](errors.md#errorcatchdefinition) | no |
| `timeout` | An ISO8601 duration string. | string | no |
| `events` | Defines the criteria by which incoming CloudEvents events are evaluated to find a match. | [ConsumeEventDefinition](./consume-event.md#consumeeventdefinition) | yes |


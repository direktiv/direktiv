# GenerateEvent State 

```yaml
- id: a
  type: generateEvent
  event:
    type: myeventtype
    source: myeventsource
    data: 
      hello: world
    datacontenttype: application/json
```

## GenerateEventStateDefinition 

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StateDefinition](./states.md) is being used. In this case it must be set to `generateEvent`. | string | yes | 
| `id` | An identifier unique within the workflow to this one state. | string | yes |
| `log` | If defined, the workflow will generate a log when it commences this state. See [StateLogging](./logging.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `metadata` | If defined, updates the instance's metadata. See [InstanceMetadata](./metadata.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `transform` | If defined, modifies the instance's data upon completing the state logic. See [StateTransforms](../instance-data/transforms.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `transition` | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, this state terminates the workflow. | string | no |
| `catch` | Defines behaviour for handling of catchable errors.  | [[]ErrorCatchDefinition](errors.md#errorcatchdefinition) | no |
| `delay` | ISO8601 duration string defining how long to hold the event before broadcasting it. | string | no |
| `event` | Defines the event to generate. | [GenerateEventDefinition](#generateeventdefinition) | yes |

## GenerateEventDefinition

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Sets the CloudEvents event type. | string | yes |
| `source` | Sets the CloudEvents event source. | string | yes |
| `data` | Defines the content of the payload for the CloudEvents event. | [Structured JQ](../instance-data/structured-jx.md) | no |
| `datacontenttype` | An RFC2046 string specifying the payload content type. | string | no |
| `context` | If defined, must evaluate to an object of key-value pairs. These will be used to define CloudEvents event context data. | [Structured JQ](../instance-data/structured-jx.md) | no |

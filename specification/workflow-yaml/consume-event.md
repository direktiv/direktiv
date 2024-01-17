# ConsumeEvent State

```yaml
- id: a
  type: consumeEvent
  timeout: PT15M
  event:
    type: com.github.pull.create
    context:
      subject: '123'
```

## ConsumeEventStateDefinition

To pause the workflow and wait until a CloudEvents event is received before proceeding, the `consumeEvent` is the simplest state that can be used. It is one of three states that can do so, along with `eventsAnd` and `eventsXor`.

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StateDefinition](./states.md) is being used. In this case it must be set to `consumeEvent`. | string | yes | 
| `id` | An identifier unique within the workflow to this one state. | string | yes |
| `log` | If defined, the workflow will generate a log when it commences this state. See [StateLogging](./logging.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `metadata` | If defined, updates the instance's metadata. See [InstanceMetadata](./metadata.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `transform` | If defined, modifies the instance's data upon completing the state logic. See [StateTransforms](../instance-data/transforms.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `transition` | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, this state terminates the workflow. | string | no |
| `catch` | Defines behaviour for handling of catchable errors.  | [[]ErrorCatchDefinition](errors.md#errorcatchdefinition) | no |
| `timeout` | An ISO8601 duration string. | string | no |
| `event` | Defines the criteria by which incoming CloudEvents events are evaluated to find a match. | [ConsumeEventDefinition](#consumeeventdefinition) | yes |

## ConsumeEventDefinition

The StartEventDefinition is a structure shared by various event-consuming states. 

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which CloudEvents events can trigger the workflow by requiring an exact match to the event's own `type` context value. | string | yes | 
| `context` | Optional key-value pairs to further restrict what events can trigger the workflow. For each pair, incoming CloudEvents context values will be checked for a match. All pairs must find a match for the event to be accepted. The "keys" must be strings that match exactly to specific context keys, but the "values" can be "glob" patterns allowing them to match a range of possible context values. | [Structured JQ](../instance-data/structured-jx.md) | no |

The received data of an event-triggered workflow is a JSON representation of all the received events stored under keys matching the events' respective type. For example, this CloudEvents event will result in the following data:

```json title="CloudEvents Event"
{
    "specversion" : "1.0",
    "type" : "com.github.pull.create",
    "source" : "https://github.com/cloudevents/spec/pull",
    "subject" : "123",
    "id" : "A234-1234-1234",
    "time" : "2018-04-05T17:31:00Z",
    "comexampleextension1" : "value",
    "comexampleothervalue" : 5,
    "datacontenttype" : "text/xml",
    "data" : "<much wow=\"xml\"/>"
}
```

```json title="Input Data"
{
  "com.github.pull.create": {
    "specversion" : "1.0",
    "type" : "com.github.pull.create",
    "source" : "https://github.com/cloudevents/spec/pull",
    "subject" : "123",
    "id" : "A234-1234-1234",
    "time" : "2018-04-05T17:31:00Z",
    "comexampleextension1" : "value",
    "comexampleothervalue" : 5,
    "datacontenttype" : "text/xml",
    "data" : "<much wow=\"xml\"/>"
  }
}
```

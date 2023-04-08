# Starts

## StartDefinition

A `StartDefinition` may be defined using one of the following, depending on the desired behaviour:

- [Starts](#starts)
  - [StartDefinition](#startdefinition)
    - [DefaultStartDefinition](#defaultstartdefinition)
    - [ScheduledStartDefinition](#scheduledstartdefinition)
    - [EventStartDefinition](#eventstartdefinition)
    - [EventsXorStartDefinition](#eventsxorstartdefinition)
    - [EventsAndStartDefinition](#eventsandstartdefinition)
    - [StartEventDefinition](#starteventdefinition)

If omitted from the workflow definition the [DefaultStartDefinition](#DefaultStartDefinition) will be used, which means the workflow will only be executed when called.

### DefaultStartDefinition

The default start definition is used for workflows that should only execute when called. This means subflows, workflows triggered by scripts, and workflows triggered manually by humans.

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StartDefinition](#startdefinition) is being used. In this case it must be set to `default`. | string | yes | 
| `state` | References a defined state's `id`. This state will be used as the entrypoint into the workflow. If left undefined, it defaults to the first state defined in the `states` list.  | string | no |

### ScheduledStartDefinition

The scheduled start definition is used for workflows that should execute at regularly defined times. 

Scheduled workflow can be manually triggered for convenience and testing. They never have input data, so accurate testing should use `{}` as input. 

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StartDefinition](#startdefinition) is being used. In this case it must be set to `scheduled`. | string | yes | 
| `state` | References a defined state's `id`. This state will be used as the entrypoint into the workflow. If left undefined, it defaults to the first state defined in the `states` list.  | string | no |
| `cron` | Defines the time(s) when the workflow should execute using a CRON expression. | string | yes |

**Example** (snippet)
```yaml
start:
  type: scheduled
  cron: '* * * * *' # Trigger a new instance every minute.
```

### EventStartDefinition 

The event start definition is used for workflows that should be executed whenever a relevant CloudEvents event is received. 

See [StartEventDefinition](#StartEventDefinition) for an explanation of the input data of event-triggered workflows.

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StartDefinition](#startdefinition) is being used. In this case it must be set to `event`. | string | yes | 
| `state` | References a defined state's `id`. This state will be used as the entrypoint into the workflow. If left undefined, it defaults to the first state defined in the `states` list.  | string | no |
| `event` | Defines what events can trigger the workflow. | [StartEventDefinition](#starteventdefinition) | yes |

### EventsXorStartDefinition 

The event "xor" start definition is used for workflows that should be executed whenever one of multiple possible CloudEvents events is received. 

See [StartEventDefinition](#starteventdefinition) for an explanation of the input data of event-triggered workflows.

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StartDefinition](#startdefinition) is being used. In this case it must be set to `eventsXor`. | string | yes | 
| `state` | References a defined state's `id`. This state will be used as the entrypoint into the workflow. If left undefined, it defaults to the first state defined in the `states` list.  | string | no |
| `events` | Defines what events can trigger the workflow.  | [[]StartEventDefinition](#starteventdefinition) | yes |

### EventsAndStartDefinition 

The event "and" start definition is used for workflows that should be executed when multiple matching CloudEvents events are received. 

See [StartEventDefinition](#starteventdefinition) for an explanation of the input data of event-triggered workflows.

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StartDefinition](#startdefinition) is being used. In this case it must be set to `eventsAnd`. | string | yes | 
| `state` | References a defined state's `id`. This state will be used as the entrypoint into the workflow. If left undefined, it defaults to the first state defined in the `states` list.  | string | no |
| `lifespan` | An ISO8601 duration string. Sets the maximum duration an event can be stored before being discarded while waiting for other events. | string | no |
| `events` | Defines what events can trigger the workflow.  | [[]StartEventDefinition](#starteventdefinition) | yes |

### StartEventDefinition

The StartEventDefinition is a structure shared by various start definitions involving events. 

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which CloudEvents events can trigger the workflow by requiring an exact match to the event's own `type` context value. | string | yes | 
| `context` | Optional key-value pairs to further restrict what events can trigger the workflow. For each pair, incoming CloudEvents context values will be checked for a match. All pairs must find a match for the event to be accepted. The "keys" are strings that match exactly to specific context keys, but the "values" can be "glob" patterns allowing them to match a range of possible context values. | object | no |

The input data of an event-triggered workflow is a JSON representation of all the received events stored under keys matching the events' respective type. For example, this CloudEvents event will result in the following input data in a workflow triggered by a single event:

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

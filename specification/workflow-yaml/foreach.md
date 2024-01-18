# Foreach State 

```yaml
- id: data
  type: noop
  transform:
    names:
    - hello
    - world
  transition: a
- id: a
  type: foreach
  array: 'jq([.names[] | {name: .}])'
  action:
    function: echo
    input: 'jq(.name)'
```

## ForeachStateDefinition

The `foreach` state is a convenient way to divide some data and then perform an action on each element in parallel. 

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StateDefinition](./states.md) is being used. In this case it must be set to `foreach`. | string | yes | 
| `id` | An identifier unique within the workflow to this one state. | string | yes |
| `log` | If defined, the workflow will generate a log when it commences this state. See [StateLogging](./logging.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `metadata` | If defined, updates the instance's metadata. See [InstanceMetadata](./metadata.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `transform` | If defined, modifies the instance's data upon completing the state logic. See [StateTransforms](../instance-data/transforms.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `transition` | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, this state terminates the workflow. | string | no |
| `catch` | Defines behaviour for handling of catchable errors.  | [[]ErrorCatchDefinition](errors.md#errorcatchdefinition) | no |
| `timeout` | ISO8601 duration string to set a non-default timeout. | string | no | 
| `array` | Selects or generates an array, from which each element will be separately acted upon. The `action.input` will be evaluated against each element in this array, rather than the usual instance data. | [Structured JQ](../instance-data/structured-jx.md) | yes | 
| `action` | Defines the action to perform. | [ActionDefinition](actions.md#actiondefinition) | yes |

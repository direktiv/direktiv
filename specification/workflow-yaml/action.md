# Action State 

```yaml
- id: a
  type: action
  action:
    function: myfunc
    input: 'jq(.x)'
```

## ActionStateDefinition 

The `action` state is the simplest and most common way to call a function or invoke a workflow to act as a subflow. See [Actions](./actions.md). 

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StateDefinition](./states.md) is being used. In this case it must be set to `action`. | string | yes | 
| `id` | An identifier unique within the workflow to this one state. | string | yes |
| `log` | If defined, the workflow will generate a log when it commences this state. See [StateLogging](./logging.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `metadata` | If defined, updates the instance's metadata. See [InstanceMetadata](./metadata.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `transform` | If defined, modifies the instance's data upon completing the state logic. See [StateTransforms](../instance-data/transforms.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `transition` | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, this state terminates the workflow. | string | no |
| `catch` | Defines behaviour for handling of catchable errors.  | [[]ErrorCatchDefinition](errors.md#errorcatchdefinition) | no |
| `timeout` | ISO8601 duration string to set a non-default timeout. | string | no | 
| `async` | If set to `true`, the workflow execution will continue without waiting for the action to return.  | boolean | no | 
| `action` | Defines the action to perform. | [ActionDefinition](actions.md#actiondefinition) | yes |

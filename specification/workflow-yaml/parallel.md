# Parallel State 

```yaml
- id: a
  type: parallel
  mode: and
  actions:
  - function: myfunc
    input: 'jq(.x)'
  - function: myfunc
    input: 'jq(.y)'
```

## ParallelStateDefinition 

The `parallel` state is an alternative to the `action` state when a workflow can perform multiple threads of logic simultaneously. The values in `return` is an array of the returns of the individual actions. In mode `or` the first response is set in the array and the other actions are set to `null`.

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StateDefinition](./states.md) is being used. In this case it must be set to `parallel`. | string | yes | 
| `id` | An identifier unique within the workflow to this one state. | string | yes |
| `log` | If defined, the workflow will generate a log when it commences this state. See [StateLogging](./logging.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `metadata` | If defined, updates the instance's metadata. See [InstanceMetadata](./metadata.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `transform` | If defined, modifies the instance's data upon completing the state logic. See [StateTransforms](../instance-data/transforms.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `transition` | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, this state terminates the workflow. | string | no |
| `catch` | Defines behaviour for handling of catchable errors.  | [[]ErrorCatchDefinition](errors.md#errorcatchdefinition) | no |
| `timeout` | ISO8601 duration string to set a non-default timeout. | string | no | 
| `mode` | If defined, must be either `and` or `or`. The default is `and`. This setting determines whether the state is considered successfully completed only if all threads have returned without error (`and`) or as soon as any single thread returns without error (`or`). | string | no | 
| `actions` | Defines the action to perform. | [[]ActionDefinition](./actions.md) | yes |

# Switch State 

```yaml
- id: a
  type: switch
  defaultTransform: 'jq(del(.x))'
  defaultTransition: b
  conditions:
  - condition: 'jq(.y == true)'
    transform: 'jq(.x)'
    transition: c
  - condition: 'jq(.z == true)'
    transform: 'jq(.x)'
    transition: d
```

## SwitchStateDefinition

To change the behaviour of a workflow based on the instance data, use a `switch` state. This state does nothing except choose between any number of different possible state transitions.

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StateDefinition](./states.md) is being used. In this case it must be set to `switch`. | string | yes | 
| `id` | An identifier unique within the workflow to this one state. | string | yes |
| `log` | If defined, the workflow will generate a log when it commences this state. See [StateLogging](./logging.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `metadata` | If defined, updates the instance's metadata. See [InstanceMetadata](./metadata.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `defaultTransform` | If defined, modifies the instance's data upon completing the state logic. But only if none of the `conditions` are met. See [StateTransforms](../instance-data/transforms.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `defaultTransition` | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, this state terminates the workflow. But only if none of the `conditions` are met. | string | no |
| `catch` | Defines behaviour for handling of catchable errors.  | [[]ErrorCatchDefinition](errors.md#errorcatchdefinition) | no |
| `conditions` | List of conditions, which are evaluated in-order until a match is found. | [[]SwitchConditionDefinition](#switchconditiondefinition) | yes |

## SwitchConditionDefinition 

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `condition` | Selects or generates the data used to determine if condition is met. The condition is considered met if the result is anything other than `null`, `false`, `{}`, `[]`, `""`, or `0`. | [Structured JQ](../instance-data/structured-jx.md) | yes | 
| `transform` | If defined, modifies the instance's data upon completing the state logic. See [StateTransforms](../instance-data/transforms.md). | [Structured JQ](../instance-data/structured-jx.md) | no | 
| `transition` | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, matching this condition terminates the workflow. | string | no | 

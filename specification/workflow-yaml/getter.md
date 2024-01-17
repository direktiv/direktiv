# Getter State

```yaml
- id: a
  type: setter
  variables:
  - key: x 
    scope: workflow
    mimeType: application/json
    value: Hello World
  transition: b
- id: b
  type: getter
  variables:
  - key: x 
    scope: workflow
```

## GetterStateDefinition

To load variables, use the `getter` state. See [Variables](../variables/variables.md).

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StateDefinition](./states.md) is being used. In this case it must be set to `getter`. | string | yes | 
| `id` | An identifier unique within the workflow to this one state. | string | yes |
| `log` | If defined, the workflow will generate a log when it commences this state. See [StateLogging](./logging.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `metadata` | If defined, updates the instance's metadata. See [InstanceMetadata](./metadata.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `transform` | If defined, modifies the instance's data upon completing the state logic. See [StateTransforms](../instance-data/transforms.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `transition` | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, this state terminates the workflow. | string | no |
| `catch` | Defines behaviour for handling of catchable errors.  | [[]ErrorCatchDefinition](errors.md#errorcatchdefinition) | no |
| `variables` | Defines variables to load. | [[]VariableGetterDefinition](#variablegetterdefinition) | yes |

## VariableGetterDefinition

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `key` | Variable name. | [Structured JQ](../instance-data/structured-jx.md) | yes |
| `scope` | Selects the scope to which the variable belongs. If undefined, defaults to `instance`. See [Variables](../variables/variables.md). | yes | no |
| `as` | Names the resulting data. If left unspecified, the `key` will be used instead. | string | no |


# Setter State

```yaml
- id: a
  type: setter
  variables:
  - key: x 
    scope: workflow
    mimeType: text/plain
    value: 'jq(.x)'
```

## SetterStateDefinition 

To create or change variables, use the `setter` state. See [Variables](../variables/variables.md).

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StateDefinition](./states.md) is being used. In this case it must be set to `setter`. | string | yes | 
| `id` | An identifier unique within the workflow to this one state. | string | yes |
| `log` | If defined, the workflow will generate a log when it commences this state. See [StateLogging](./logging.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `metadata` | If defined, updates the instance's metadata. See [InstanceMetadata](./metadata.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `transform` | If defined, modifies the instance's data upon completing the state logic. See [StateTransforms](../instance-data/transforms.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `transition` | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, this state terminates the workflow. | string | no |
| `catch` | Defines behaviour for handling of catchable errors.  | [[]ErrorCatchDefinition](errors.md#errorcatchdefinition) | no |
| `variables` | Defines variables to push. | [[]VariableSetterDefinition](#variablesetterdefinition) | yes |

## VariableSetterDefinition

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `key` | Variable name. | [Structured JQ](../instance-data/structured-jx.md) | yes |
| `scope` | Selects the scope to which the variable belongs. If undefined, defaults to `instance`. See [Variables](../variables/variables.md). | yes | no |
| `mimeType` | Store a MIME type with the variable. If left undefined, it will default to `application/json`. Two specific MIME types cause this state to behave differently: `text/plain` and `application/octet-stream`. If the `value` evaluates to a JSON string the MIME type is `text/plain`, that string will be stored in plaintext (without JSON quotes and escapes). If if the `value` is a JSON string containing base64 encoded data and the MIME type is `application/octet-stream`, the base64 data will be decoded and stored as binary data. | [Structured JQ](../instance-data/structured-jx.md) | no |
| `value` | Select or generate the data to store.  | [Structured JQ](../instance-data/structured-jx.md) | yes |
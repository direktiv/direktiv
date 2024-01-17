# Validate State 

```yaml
- id: a
  type: validate
  schema:
    title: Files
    type: object
    properties:
      firstname:
        type: string
        description: Your first name
        title: First Name
```

## ValidateStateDefinition

Since workflows receive external input it may be necessary to check that instance data is valid. The `validate` state exists for this purpose. If this state is the first state in the flow the UI will generate a input form based on the specification.

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StateDefinition](./states.md) is being used. In this case it must be set to `validate`. | string | yes | 
| `id` | An identifier unique within the workflow to this one state. | string | yes |
| `log` | If defined, the workflow will generate a log when it commences this state. See [StateLogging](./logging.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `metadata` | If defined, updates the instance's metadata. See [InstanceMetadata](./metadata.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `transform` | If defined, modifies the instance's data upon completing the state logic. See [StateTransforms](../instance-data/transforms.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `transition` | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, this state terminates the workflow. | string | no |
| `catch` | Defines behaviour for handling of catchable errors.  | [[]ErrorCatchDefinition](errors.md#errorcatchdefinition) | no |
| `subject` | Selects or generates the data that will be compared to the `schema`. If undefined, it will be default to `'jq(.)'`. | [Structured JQ](../instance-data/structured-jx.md) | no |
| `schema` | A YAMLified representation of a JSON Schema that defines whether the `subject` is considered valid. | object | yes |

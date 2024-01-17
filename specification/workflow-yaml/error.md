# Error State

```yaml
- id: a
  type: error
  error: badinput
  message: 'Missing or invalid value for required input.'
```

## ErrorStateDefinition 

When workflow logic end up in a failure mode, the `error` state can be used to mark the instance as failed. This allows the instance to report what went wrong to the caller, which may then be handled or reported appropriately.

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StateDefinition](./states.md) is being used. In this case it must be set to `error`. | string | yes | 
| `id` | An identifier unique within the workflow to this one state. | string | yes |
| `log` | If defined, the workflow will generate a log when it commences this state. See [StateLogging](./logging.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `metadata` | If defined, updates the instance's metadata. See [InstanceMetadata](./metadata.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `transform` | If defined, modifies the instance's data upon completing the state logic. See [StateTransforms](../instance-data/transforms.md). | [Structured JQ](../instance-data/structured-jx.md) | no |
| `transition` | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, this state terminates the workflow. | string | no |
| `catch` | Defines behaviour for handling of catchable errors.  | [[]ErrorCatchDefinition](errors.md#errorcatchdefinition) | no |
| `error` | A short descriptive error code that can be caught by a parent workflow. | string | yes |
| `message` | Generates a more detailed message or object that can contain instance data, to provide more information for users. | [Structured JQ](../instance-data/structured-jx.md) | yes |

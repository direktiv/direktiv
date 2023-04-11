# Transforms

Whenever an instance finishes executing a state there is an opportunity to perform a Transform. Usually with a field called `transform`, but sometimes in other forms. The `switch` state also has a `defaultTransform`, for example. 

All transforms use [structured jx](./structured-jx.md), giving you powerful options to enrich, sanitize, or modify instance data. All transforms must produce output that remains valid [instance data](./instance-data.md), otherwise an error will be thrown: `direktiv.jq.notObject`.

## Examples

Here are some common use-case helpful examples of transforms.

### Completely Replacing Instance Data

```json title="Instance Data Before Transform"
{
  "msg": "Hello, world!
}
```

```yaml title="Transform Snippet"
- id: snippet
  type: noop
  transform: 
    x: 5
```

```json title="Instance Data After Transform"
{
  "x": 5
}
```

### Replacing A Subset Of Instance Data 

```json title="Instance Data Before Transform"
{
  "a": 1,
  "b": 2,
  "c": 3
}
```

```yaml title="Transform Snippet"
- id: snippet
  type: noop
  transform: 'jq(.a = 5 | .b = 6)'
```

```json title="Instance Data After Transform"
{
  "a": 5,
  "b": 6,
  "c": 3
}
```

### Deleteing A Subset of Instance Data 

```json title="Instance Data Before Transform"
{
  "a": 1,
  "b": 2,
  "c": 3
}
```

```yaml title="Transform Snippet"
- id: snippet
  type: noop
  transform: 'jq(del(.a) | del(.b))'
```

```json title="Instance Data After Transform"
{
  "c": 3
}
```

### Adding A New Value.

```json title="Instance Data Before Transform"
{
  "a": 1
}
```

```yaml title="Transform Snippet"
- id: snippet
  type: noop
  transform: 'jq(.b = 2)'
```

```json title="Instance Data After Transform"
{
  "a": 1,
  "b": 2
}
```

### Renaming A Subset of Instance Data

```json title="Instance Data Before Transform"
{
  "a": 1,
  "b": 2,
  "c": 3
}
```

```yaml title="Transform Snippet"
- id: snippet
  type: noop
  transform: 'jq(.x = .a | del(.a))'
```

```json title="Instance Data After Transform"
{
  "b": 2,
  "c": 3,
  "x": 1
}
```

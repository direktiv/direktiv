# Transforms

Whenever an instance finishes executing a state there is an opportunity to perform a Transform. Usually with a field called `transform`, but sometimes in other forms. The `switch` state also has a `defaultTransform`, for example. 

All transforms use [structured jx](./structured-jx.md), giving you powerful options to enrich, sanitize, or modify instance data. All transforms must produce output that remains valid [instance data](./instance-data.md), otherwise an error will be thrown. // TODO: what error, specifically?

## Examples

Here are some common use-case helpful examples of transforms.

### Completely Replacing Instance Data

**Instance Data Before Transform**

```json
{
	"msg": "Hello, world!
}
```

**Transform Snippet**

```yaml
- id: snippet
  type: noop
  transform: 
    x: 5
```

**Instance Data After Transform**

```json
{
	"x": 5
}
```

### Replacing A Subset Of Instance Data 

**Instance Data Before Transform**

```json
{
	"a": 1,
	"b": 2,
	"c": 3
}
```

**Transform Snippet**

```yaml
- id: snippet
  type: noop
  transform: 'jq(.a = 5 | .b = 6)'
```

**Instance Data After Transform**

```json
{
	"a": 5,
	"b": 6,
	"c": 3
}
```

### Deleteing A Subset of Instance Data 

**Instance Data Before Transform**

```json
{
	"a": 1,
	"b": 2,
	"c": 3
}
```

**Transform Snippet**

```yaml
- id: snippet
  type: noop
  transform: 'jq(del(.a) | del(.b))'
```

**Instance Data After Transform**

```json
{
	"c": 3
}
```

### Adding A New Value.

**Instance Data Before Transform**

```json
{
	"a": 1
}
```

**Transform Snippet**

```yaml
- id: snippet
  type: noop
  transform: 'jq(.b = 2)'
```

**Instance Data After Transform**

```json
{
	"a": 1,
	"b": 2
}
```

### Renaming A Subset of Instance Data

**Instance Data Before Transform**

```json
{
	"a": 1,
	"b": 2,
	"c": 3
}
```

**Transform Snippet**

```yaml
- id: snippet
  type: noop
  transform: 'jq(.x = .a | del(.a))'
```

**Instance Data After Transform**

```json
{
	"b": 2,
	"c": 3,
	"x": 1
}
```

# Structured JX

Many fields of the workflow definition are described as "Structured JX". That's a name we use for fields that support complex and powerful query logic that we'll describe in greater detail here.

## JQ 

Since [instance data](./instance-data.md) is represented as JSON, the most natural way to work with that data is with the powerful JSON query language called jq. 

Whenever a string appears within a Structured JX field that includes `jq(...)`, everything between the brackets is evaluated as a jq query against the instance data. Then the entire `jq(...)` part is replaced by the results of that query. 

> Note: YAML allows for strings without quotation marks, but this should be avoided when using Structured JX. The characters in the queries will commonly be interpreted in unintended ways by the YAML parser.

If the `jq(...)` part constitutes the entirety of the string then the entire string is replaced by whatever data type was returned. If not, the results are marshalled into a JSON string and substituted into the parent string. 

The one exception to this rule is if the returned data type is a string, in which case it is substituted as-is without marshalling into JSON. This enables you to build strings without filling them with quotation marks.

### Example 1

**Instance Data**

```json
{
	"a": [1, 2, 3]
}
```

**Structured JX**

```yaml
'jq(.a)'
```

**Evaluated Result**

```json
[1, 2, 3]
```

### Example 2

**Instance Data**

```json
{
	"a": [1, 2, 3]
}
```

**Structured JX**

```yaml
'a: jq(.a)'
```

**Evaluated Result**

```json
"a: [1, 2, 3]"
```

### Example 3

**Instance Data**

```json
{
	"a": "hello"
}
```

**Structured JX**

```yaml
'a: jq(.a)'
```

**Evaluated Result**

```json
"a: hello"
```

## JS 

JQ isn't the only option available to interact with the instance data. Javascript is also supported using `js(...)` in a very similar way. Entire Javascript scripts can be embedded in strings within Structured JX.

> Note: YAML supports several ways of including large or multi-line strings. But each of these ways is treated a little bit differently by the YAML parser. To preserve newlines, we recommend using the `|` form. With Javascript this often necessary. 

When writing scripts this way, the instance data is copied and exposed to the script in an object called `data`. 

### Example 1

**JQ**

```yaml
transform: 'jq({x: 5})'
```

**Analogous Javascript**

```yaml
transform: |
  js(
    items = new Object()
    items.x = 5
    return items
  )
```

## Example 2

**JQ**

```yaml
transform: 'jq({x: .a})'
```

**Analogous Javascript**

```yaml
transform: |
  js(
    items = new Object()
    items.x = data['a']
    return items
  )
```

## YAML

So far we've seen how you can use jq or Javascript to produce a value for your Structured JX field, but it's also possible to use neither, or both. 

The "Structured" part of Structured JX is so named because you don't have to provide a single string. You can provide any type of data you like. The entirety of what is provided will be converted from its YAML representation to a JSON representation. And then every field within will be searched recursively for embedded jq/Javascript. 

### Example

**Instance Data Before Transform**

```json
{
	"a": [1, 2, 3]
}
```

**Transform**

```yaml
tranform:
  x: 'jq(.a)'
  y: |
    js(
	  var output = data['a'].map((x) => {return ++x;})
      return output
	)
  z: 5
  listA: ["a", "b", "c"]
  listB:
  - d
  - e
  - f
  obj:
    i: 10
	j: 'jq(.a[2])'
```

**Evaluated Result**

```json
{
	"listA": ["a", "b", "c"],
	"listB": ["d", "e", "f"],
	"obj": {
		"i": 10,
		"j": 3
	},
	"x": [1, 2, 3],
	"y": [2, 3, 4],
	"z": 5
}
```

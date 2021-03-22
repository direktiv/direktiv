# Hello World

A simple `noop` state that transforms the output to say 'Hello World!'.

## Workflow

```yaml
id: helloworld 
states:
- id: hello
  type: noop
  transform: '{ result: "Hello World!" }'
```

## Input

```json
null
```

## Output

```json
{
	"result": "Hello World!"
}
```
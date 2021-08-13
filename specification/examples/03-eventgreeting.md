# Event Greeting

## Workflow

```yaml
id: eventgreeting
functions: 
- id: greetingFunction
  type: reusable
  image: vorteil/greeting
states:
- id: Begin
  type: consumeEvent
  event:
    type: greetingEventType
    context: 
      source: greetingEventSource
  transition: Greet
- id: Greet
  type: action
  action:
    function: greetingFunction
    input: '{ name: .greet.name }'
  transform: '{ greeting: .return.greeting }'
```

## Input 

```json
{
    "specversion" : "1.0",
    "type" : "greetingEventType",
    "source" : "greetingEventSource",
    "data" : {
      "greet": {
          "name": "Trent"
        }
    }
}
```

## Output

```json
{ 
	"greeting": "Welcome to Direktiv, Trent!"
}
```

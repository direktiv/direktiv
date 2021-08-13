# Greeting

A simple action that uses the docker container `vorteil/greeting`. Which takes a person object as input and outputs a greeting message back.

## Workflow 

```yaml
id: greeting
functions: 
- id: greetingFunction
  type: reusable
  image: vorteil/greeting
states:
- id: Greet
  type: action
  action:
    function: greetingFunction
    input: '{ name: .person.name }'
  transform: '{ greeting: .return.greeting }'
```

## Input 

```json
{
  "person": {
    "name": "Trent"
  }
}
```

## Output 

```json
{
   "greeting":  "Welcome to Direktiv, Trent!"
}
```
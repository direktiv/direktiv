# Transitions

The ability to string a number of different operations together is a fundamental part of workflows. In this article you'll learn about transitions, transforms, and `jq`. 

## Demo

```yaml
id: transitioner
states:
- id: a
  type: noop
  transform: '{
    "number": 2,
    "objects": [{
      "k1": "v1"
    }]
  }'
  transition: b
- id: b
  type: noop
  transform: '.multiplier = 10'
  transition: c
- id: c
  type: noop
  transform: '.result = .multiplier * .number | del(.multiplier, .number)'
  transition: d
- id: d
  type: noop
  transform: '.objects[0]'
```

### Input

```json
{}
```

### Output

```json 
{
  "k1": "v1"
}
```

### Logs

```
[10:10:30] Beginning workflow triggered by API.
[10:10:30] Running state logic -- a:1 (noop)
[10:10:30] State data:
{}
[10:10:30] Transforming state data.
[10:10:30] Transitioning to next state: b (1).
[10:10:30] Running state logic -- b:2 (noop)
[10:10:30] State data:
{
  "number": 2,
  "objects": [
    {
      "k1": "v1"
    }
  ]
}
[10:10:30] Transforming state data.
[10:10:30] Transitioning to next state: c (2).
[10:10:30] Running state logic -- c:3 (noop)
[10:10:30] State data:
{
  "multiplier": 10,
  "number": 2,
  "objects": [
    {
      "k1": "v1"
    }
  ]
}
[10:10:30] Transforming state data.
[10:10:30] Transitioning to next state: d (3).
[10:10:30] Running state logic -- d:4 (noop)
[10:10:30] State data:
{
  "objects": [
    {
      "k1": "v1"
    }
  ],
  "result": 20
}
[10:10:30] Transforming state data.
[10:10:30] Workflow completed.
```

## Transitions

More than one state can be defined in a workflow definition. Each begins under the `states` field and multiple states can be differentiated by looking for the dash symbol that denotes a new object in the list of states. In the demo there are four separate states:

### State 'a'

```yaml
- id: a
  type: noop
  transform: '{
    "number": 2,
    "objects": [{
      "k1": "v1"
    }]
  }'
  transition: b
```

### State 'b'

```yaml
- id: b
  type: noop
  transform: '.multiplier = 10'
  transition: c
```

### State 'c'

```yaml
- id: c
  type: noop
  transform: '.result = .multiplier * .number | del(.multiplier, .number)'
  transition: d
```

### State 'd'

```yaml
- id: d
  type: noop
  transform: '.objects[0]'
```

We've only got Noop States here, but most state types may optionally have a `transition` field, with a reference to the identifier for a state in the workflow definition. After a state finishes running Direktiv uses this field to figure out whether the instance has reached its end or not. If a transition to another state is defined the instance will continue on to that state.

In this demo four Noop States are defined in a simple sequence that goes `a → b → c → d`. The instance data for each state is inherited from its predecessor, which is why it can be helpful to use Transforms.

## Transforms & JQ

Every workflow instance always has something called the "Instance Data", which is a JSON object that is used to pass data around. Almost everywhere a `transition` can happen in a workflow definition a `transform` can also happen allowing the author to filter, enrich, or otherwise modify the instance data.

The `transform` field can contain a valid `jq` command, which will be applied to the existing instance data to generate a new JSON object that will entirely replace it. Note that only a JSON **object** will be considered a valid output from this `jq` command: `jq` is capable of outputting primitives and arrays, but these are not acceptable output for a `transform`. 

Because the Noop State logs its instance data before applying its `transform` & `transition` we can follow the results of these transforms throughout the demo.

### Input

```json
{}
```

### First Transform

The first transform defines a completely new JSON object.

**Command**

```yaml
  transform: '{
    "number": 2,
    "objects": [{
      "k1": "v1"
    }]
  }'
```

**Resulting Instance Data**

```json
{
  "number": 2,
  "objects": [
    {
      "k1": "v1"
    }
  ]
}
```

### Second Transform

The second transform enriches the existing instance data by adding a new field to it.

**Command**

```yaml
  transform: '.multiplier = 10'
```

**Resulting Instance Data**

```json
{
  "multiplier": 10,
  "number": 2,
  "objects": [
    {
      "k1": "v1"
    }
  ]
}
```

### Third Transform 

The third transform multiplies two fields to produce a new field, then pipes the results into another command that deletes two fields.

**Command**

```yaml
  transform: '.result = .multiplier * .number | del(.multiplier, .number)'
```

**Resulting Instance Data**

```json
{
  "objects": [
    {
      "k1": "v1"
    }
  ],
  "result": 20
}
```

### Fourth Transform

The fourth transform selects a child object nested within the instance data and makes that into the new instance data.

**Command**

```yaml
  transform: '.objects[0]'
```

**Resulting Instance Data**

```json
{
  "k1": "v1"
}
```


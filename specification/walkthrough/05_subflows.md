# Subflows

Just like scripting or programming, with Direktiv it's possible to organize your logic into reusable modules. Anytime a workflow is invoked by another we call it a subflow. In this article you'll learn about namespaces, subflows, and instance data validation. 

## Demo 

For a subflow demonstration we need to define multiple workflows. 

### 1st Workflow Definition

```yaml
id: notifier
functions:
- id: httprequest
  type: reusable
  image: vorteil/request
states:
- id: validateInput
  type: validate
  schema:
    type: object
    required:
    - contact
    - payload
    additionalProperties: false
    properties:
      contact:
        type: string 
      payload:
        type: string
  transition: notify
- id: notify
  type: action 
  action:
    function: httprequest
    input: '{
      "method": "POST",
      "url": "https://jsonplaceholder.typicode.com/todos/1",
      "body": .input
    }'
  transition: checkResults
- id: checkResults
  type: switch
  conditions:
  - condition: '.warnings'
    transition: throw
- id: throw
  type: error
  error: notification.lint
  message: 'lint errors: %s'
  args:
  - '.warnings'
```

### 2nd Workflow Definition

```yaml
id: worker
functions: 
- id: httprequest
  type: reusable
  image: vorteil/request
states:
- id: do
  type: action
  action:
    function: httprequest
    input: '{
      "method": "POST",
      "url": "https://jsonplaceholder.typicode.com/todos/1",
      "body": .input
    }'
  transition: notify
  transform: 'del(.return) | .contact = "Alan"'
- id: notify
  type: action 
  action:
    workflow: notifier
    input: '{ contact: .contact, payload: .input }'
```

### Input

```
Hello, world!
```

### Output

TODO

## Namespaces

Before learning about subflows you'll need to know what a "Namespace" is. Direktiv organizes everything into namespaces. Think of them a bit like a folder. 

All of your workflow definitions exist within a namespace, and any instances spawned from those definitions exist within that namespace as well. Any secrets or registries you've set up apply only within the namespace (more on these in a later article). Some limitations are applied on a namespace-level. And frontends may piggyback on the namespaces to handle permissions and multi-tennancy.

Workflow identifiers are unique within a namespace, which allows them to be referenced as subflows.

## Subflows

Anywhere an "Action" appears in a workflow definition either an Isolate or a Subflow can be run. Here's an example of what a subflow call could look like:

```yaml
id: httpget
states:
- id: getter 
  type: action
  action:
    workflow: myworkflow
    input: '{
      "method": "GET",
      "url": "https://jsonplaceholder.typicode.com/todos/1",
    }'
```

Nothing special needs to be done when writing a workflow definition that's intended for use as a subflow. Input is treated exactly the same way as if the workflow was directly invoked with the API, and output is merged into the caller's instance data exactly the same way as if the subflow was an Isolate. 

## Validate State 

So far we've never demonstrated any way to validate external inputs. Validation is optional, but it can be important for preventing bugs in your workflows caused by unexpected data. Unexpected data could occur in many ways: bad workflow input, bad events, bad Isolate results, or mistakes in your own transforms. Using the Validate State Direktiv can detect these issues and throw an error instead of proceeding.

From the example in the demo, here's what a Validate State definition might look like:

```yaml
- id: validateInput
  type: validate
  schema:
    type: object
    required:
    - contact
    - payload
    additionalProperties: false
    properties:
      contact:
        type: string 
      payload:
        type: string
  transition: notify
```

Unlike any other state, the fields for a Validate State are not fixed. Direktiv takes everything under the `schema` field and converts it into an equivalent [JSON Schema](https://json-schema.org/), which it then uses to validate the instance data. Converting JSON Schema from JSON to YAML is straight-forward. Here's the equivalent JSON for the YAML schema in the example above:

```json
{
  "type": "object",
  "required": ["contact", "payload"],
  "additionalProperties": false,
  "properties": {
    "contact": {
      "type": "string"
    },
    "payload": {
      "type": "string"
    }
  }
}
```

There are also many tools online that can convert JSON to YAML for you. 

If the instance data fails its validation Direktiv will throw a `direktiv.schema.failed` error, which will terminate the workflow unless an appropriate error catcher is defined (more on error handling in a later article).

## Error State

Speaking of errors, there's another new state in this demo example: the Error State. The Error State is only really useful in the context of subflows. 

```yaml
- id: throw
  type: error
  error: notification.lint
  message: 'lint errors: %s'
  args:
  - '.warnings'
```

If an instance executes an Error State it will store the custom-defined error and mark the instance as failed after the instance terminates. 

The Error State has an optional `transition` field just like every other state, which might surprise you. That's because the Error State won't actually cause the workflow to terminate like you might expect. This is to allow the workflow to perform any cleanup, rollback, or recovery logic. The error will still be reported when the instance does finally finish. 

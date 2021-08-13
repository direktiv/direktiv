# Conditional Transitions 

Oftentimes a workflow needs to be a little bit smarter than an immutable sequence of states. That's when you might need conditional transitions. In this article you'll learn about instance input data, the Switch State, and how to define loops.

## Demo 

```yaml
id: multiposter
functions:
- id: httprequest
  type: reusable
  image: vorteil/request
states:
- id: ifelse
  type: switch
  conditions:
  - condition: '.names'
    transition: poster
- id: poster
  type: action
  action:
    function: httprequest
    input: '{
      "method": "POST",
      "url": "https://jsonplaceholder.typicode.com/posts",
      "body": {
        "name": .names[0]	
      }
    }'
  transform: 'del(.names[0])'
  transition: ifelse
```

### Input 

```json
{
  "names": [
    "Alan",
    "Jon",
    "Trent"
  ]
}
```

### Output

```json
{
  "names": [],
  "return": {
    "id": 101
  }
}
```

## Instance Input

Workflows can be invoked with input data that will be available as instance data. There are a few ifs-and-buts that apply to input data, but it's not that complicated.   

### Input JSON Object

If the input is a JSON object it will become the instance data.

**Input Data**

```json
{
  "key": "value"	
}
```

**Instance Data**

```json
{
  "key": "value"	
}
```

### Input JSON Non-Object

If the input is valid JSON but **not** an object it will be stored under `"input"`.

**Input Data**

```json
[1, 2, 3]
```

**Instance Data**

```json
{
  "input": [1, 2, 3]	
}
```

### Input Non-JSON

Finally, if the input is not valid JSON it will be base64 encoded into a string and then stored under `"input"`.

```
Hello, world!
```

**Instance Data**

```json
{
  "input": "SGVsbG8sIHdvcmxkIQ=="	
}
```

## Switch State

The Switch State can make decisions about where to transition to next based on the instance data by evaluating a number of `jq` expressions and checking the results. Here's an example switch state definition:

```yaml
- id: ifelse
  type: switch
  conditions:
  - condition: '.person.age > 18'
    transition: accept
    #transform:
  - condition: '.person.age != null'
    transition: reject
    #transform:
  defaultTransition: failure
  #defaultTransform: 
```

Each of the `conditions` will be evaluated in the order it appears by running the `jq` command in `condition`. Any result other than `null`, `false`, `{}`, `[]`, `""`, or `0` will cause the condition to be considered a successful match. If no conditions match the default transition will be used. 

In the demo example the switch state will transition to `poster` until the list of names is empty, at which point the workflow will end.

## Other Conditional Transitions

The Switch State is not the only way to do conditional transitions. The EventsXor state also transitions conditionally based on which CloudEvent was received. All states can also define handlers for catching various types of errors. Both of these will be discussed in a later article.

## Loops

By transitioning to a state that has already happened it's possible to create loops in workflow instances. In this demo we've got a type of range loop, iterating over the contents of an array. Direktiv sets limits for the number of transitions an instance can make in order to protect itself from infinitely-looping workflows. Consider some of the following alternatives if you want to have a loop:

### Foreach State

For range loops like the one in this demo there's another state called a Foreach State that simplifies the logic and splits up a data set to run many actions in parallel without doing lots of transitions. A fuller explanation of the Foreach State will be discussed in a later article, but here's an equivalent workflow definition to the demo if you're curious:

```yaml
id: multiposter
functions:
- id: httprequest
  type: reusable
  image: vorteil/request
states:
- id: poster
  type: foreach
  array: '.names[] | { name: . }'
  action:
    function: httprequest
    input: '{
      "method": "POST",
      "url": "https://jsonplaceholder.typicode.com/posts",
      "body": {
        "name": .	
      }
    }'
  transform: 'del(.names) | .names = []'
```

### Retries

One obvious use for loops is to retry some logic if an error occurs, but there's no need to design looping workflow because Direktiv has configurable error catching & retrying available on every state. This will be discussed in a later article.

### Isolates

For large data sets or logic that could needs to loop many times it's generally better to custom-write an Isolate function that performs all of the computation. Writing custom Isolates is discussed in another article.

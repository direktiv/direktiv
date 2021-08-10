# Introduction to Isolates

Workflows wouldn't be very powerful if they were limited to just the predefined states. That's why Direktiv can run "Isolates": tiny Virtual Machines based on containers and powered by Vorteil.io that are built on demand and scrapped as soon as they've finished. In this article you'll learn about the Action State and get an introduction to Isolates.

## Demo

```yaml
id: httpget
functions:
- id: httprequest
  type: reusable
  image: vorteil/request
states:
- id: getter 
  type: action
  action:
    function: httprequest
    input: '{
      "method": "GET",
      "url": "https://jsonplaceholder.typicode.com/todos/1",
    }'
```

This workflow will use the Docker container at https://hub.docker.com/r/vorteil/request to perform a GET request and return the results to the instance data. 

Not just any Docker container will work as an Isolate, but it isn't difficult to make one compatible. We'll discuss that later.

Run this workflow. Leave the Workflow Input empty for now. You should see something like the following:

### Input

```json
{}
```

### Output

```json
{
  "return": {
    "userId": 1,
    "id": 1,
    "title": "delectus aut autem",
    "completed": false
  }
}
```

The JSON structure under `"return"` is the object returned by the GET request.

## Introduction to Isolates

Isolate is just a fancy term we use when we run a tiny virtual machine on Vorteil.io. Direktiv grabs a Docker container from an available Docker Registry: hub.docker.com unless custom registries are defined (more on that later). It then converts this container into a Vorteil virtual machine to run as a "function". If the container handles input and output according to our Isolate requirements it can do just about anything (more on our Isolate requirements later as well).

### Function Definitions

```yaml
functions:
- id: httprequest
  type: reusable
  image: vorteil/request
```

To use an Isolate it must first be defined at the top of the workflow definition. Each function definition needs an identifier that must be unique within the workflow definition, and an `image` that references a Docker container to use.

## Action State

```yaml
- id: getter 
  type: action
  action:
    function: httprequest
    input: '{
      "method": "GET",
      "url": "https://jsonplaceholder.typicode.com/todos/1",
    }'
```

Like all other states, the Action State requires an `id` and `type` field identifying it as such. But the great thing about the Action State is its ability to run user-made logic in the form of "Isolates". 

The `function` field must reference one of the `functions` defined in the workflow definition. In this example we're using `vorteil/request`, which is a simple container that performs a HTTP request and returns the results. We use a `jq` command specified in the `input` field to generate the input for the Isolate.

Once the Isolate has completed its task in the Action State the results are stored in the instance data under the `"return"` field.

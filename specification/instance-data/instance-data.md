# Instance Data

Every workflow instance has its own **instance data**, which is data that is exclusively accessible to the instance, and only changeable by the instance. 

## JSON Objects

Instance data is represented in JSON form, and is always a valid JSON **object**. This detail is important because it means that not everything that is valid JSON can be valid instance data. 

This is valid instance data:

```json
{}
```

So is this:

```json
{
  "list": [1, 2, 3]
}
```

And this:

```json
{
  "a": 5,
  "b": "6",
  "c": true,
  "d": {
    "list": [7, "8"]
  }
}
```

But this are not valid instance data, even though it is valid JSON:

```json
true
```

Neither is this:

```json
"Hello, world!"
```

Nor is this:

```json
[{
  "a": 5
}]
```

Another way of looking at it: it's not valid instance data unless it's valid JSON beginning with `{` and ending with `}`.

## Size Limit

The size of instance data is measured in terms of the length (in bytes) of its JSON representation. For technical reasons, there is an enforced upper limit allowed for this maximum size. This limit can vary according to the specific configuration of a Direktiv installation, but the default is 128 MiB.

## Lifecycle

The starting value for an instance's data is set based on what triggered the workflow to spawn a new instance. See [Workflow Input](./input.md).

Afterwards, the instance may manipulate the data in predictable ways according to the instructions in the workflow definition. The main way to change instance data is through [Transforms](./transforms.md). 

Other operations can also contribute to the instance data. Actions may return results, error handling may save error information, event listeners save received events, and variable getters can retrieve data saved elsewhere and add it to the instance data.

After the final operation of an instance is executed the instance data becomes the instance's output data. See [Instance Output](./output.md). Output data is viewable by the API, and is also returned to the caller if the instance was executed as a subflow to another workflow.

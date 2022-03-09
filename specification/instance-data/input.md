# Workflow Input

When a workflow is triggered and spawns a new instance it may do so with some starting value for its [instance data](./instance-data.md). Here's everything you need to know about workflow input. 

## API / Subflow

If a workflow is invoked directly, either through the API or as a subflow to another instance, its input is passed in as-is. 

So if you call the workflow with the following input:

```json
{
	"msg": "Hello, world!"
}
```

Then the instance data for the workflow will, be the same:

```json
{
	"msg": "Hello, world!"
}
```

That is, unless the input data isn't a valid JSON object. If the input is valid JSON but not an object, as in the following example, it is wrapped within an object automatically under the property `.input`. 

So this input:

```json
[1, 2, 3]
```

Becomes:

```json
{
	"input": [1, 2, 3]
}
```

If the input data isn't valid JSON at all, it is treated as binary data. Binary data is converted into a base64 encoded string and passed into the instance the same way as above.

This input:

```
Hello, world!
```

Becomes:

```json
{
	"input": "SGVsbG8sIHdvcmxkIQo="
}
```

This treatment of binary data allows workflows to handle non-JSON inputs. Common examples include XML and form data. Just use a function to extract the information needed from these other formats and convert them to JSON.

One thing to keep in mind that might trip you up: if you provide no input data whatsoever that's not valid JSON. It is valid binary data, which means this input:

```
```

Becomes:

```json
{
	"input": ""
}
```

An empty string is a valid base64 representation of zero bytes.

## CRON

By their nature, scheduled workflows have empty input. They will always be:

```json
{}
```

This doesn't mean they have to do exactly the same thing each time, it just means you need to get a little creative. For example, begin your workflow by loading data from variables or by using an action that grabs data from an external source.

## CloudEvents Events

Workflows that are triggered by receiving one or more events will include the received event(s) in their input data. Each received event will appear in the instance data under a property with the same value as their event type, to allow workflows to distinguish between events.

For an instance triggered with the following event:

```json
{
    "specversion" : "1.0",
    "type" : "com.github.pull.create",
    "source" : "https://github.com/cloudevents/spec/pull",
    "subject" : "123",
    "id" : "A234-1234-1234",
    "time" : "2018-04-05T17:31:00Z",
    "comexampleextension1" : "value",
    "comexampleothervalue" : 5,
    "datacontenttype" : "text/xml",
    "data" : "<much wow=\"xml\"/>"
}
```

The instance input data becomes:

```json
{
	"com.github.pull.create": {
		"specversion" : "1.0",
		"type" : "com.github.pull.create",
		"source" : "https://github.com/cloudevents/spec/pull",
		"subject" : "123",
		"id" : "A234-1234-1234",
		"time" : "2018-04-05T17:31:00Z",
		"comexampleextension1" : "value",
		"comexampleothervalue" : 5,
		"datacontenttype" : "text/xml",
		"data" : "<much wow=\"xml\"/>"
	}
}
```

If an event's payload is JSON it should be directly addressable, rather than being embedded within a string.

## Large Inputs

Like instance data, input data has size limits. These size limits are usually the same, but not necessarily. This will vary according to the configuration of each Direktiv installation, and is usually about 32 MiB. 

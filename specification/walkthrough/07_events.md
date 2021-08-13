# Events

Direktiv has built-in support for CloudEvents, which can be a great way to interact with workflows. In this article you'll learn about events.

## Demo 

```yaml 
id: notifier
start:
  type: event
  event: 
    type: com.github.pull.create
    filters:
      source: "https://github.com/cloudevents/spec/pull"
functions:
- id: httprequest
  type: reusable
  image: vorteil/request
states:
- id: notify
  type: action 
  action:
    function: httprequest
    input: '{
      "method": "POST",
      "url": "https://jsonplaceholder.typicode.com/todos/1",
      "body": ."com.github.pull.create"
    }'
```

## CloudEvents 

[CloudEvents](https://cloudevents.io/) are specification for describing event data in a common way. They're JSON objects with a number of required fields, some optional fields, and a payload. Here's an example CloudEvent:

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

CloudEvents can be sent via the API to a namespace, to be handled by any number of interested receivers on that namespace. 

## Start Types

The most common use for events in Direktiv is to have external services generate CloudEvents and send them to Direktiv to trigger your workflows. But to make your workflows trigger on an event you need to register the workflow's interest in the event by adding the appropriate start type to your workflow definition:

```yaml
start:
  type: event
  event: 
    type: com.github.pull.create
    filters:
      source: "https://github.com/cloudevents/spec/pull"
```

In this example a new instance will be created from our workflow whenever a cloudevent is received that has the matching `type` and `source` values.

Two other event-based start types exist in Direktiv: the `eventsXor`, and the `eventsAnd`. 

The `eventsXor` registers an interest in multiple events and will trigger a new instance as soon as any one of them is received. The `eventsAnd` also registers an interest in multiple events, but will only trigger once all have been received. 

## Event Payloads

Whenever an event is received its payload will be added to the instance data under a field with the same name as the event "type". This allows for a uniform approach to accepting events that supports single events, eventsXor, and eventsAnd. 

Like direct input, the payload will be treated as nested JSON if possible, but will be base64 encoded if it's some other content type. So, for the example workflow and event above, the instance data before running the first state would be the following:

```json
{
  "com.github.pull.create": "PG11Y2ggd293PSJ4bWwiLz4="
}
```

## Instances Waiting for Events 

Triggering workflows is not the only thing you can do with events. Workflows can be constructed to run some logic and then wait for an event before proceeding. Like the event-based start types, there are three event consuming states: `consumeEvent`, `eventsXor`, and `eventsAnd`. Here's an example of what a ConsumeEvent State could look like:

```yaml
- id: waitEvent
  type: consumeEvent 
  event: 
    type: com.github.pull.create 
    context:
      source: "https://github.com/cloudevents/spec/pull"
      repository: '{{ .repo }}'
  timeout: PT5M 
  transform: '."com.github.pull.create"'
  transition: nextState
```

### Timeouts

It's rarely a good idea to leave a workflow waiting indefinitely. Direktiv allows you to define timeouts in ISO8601 format when waiting on an event. If the state is not ready to proceed before the timeout has elapsed an error will be thrown. It's possible to catch this error, but that's for a later article.

The `timeout` field is not required, but Direktiv caps the maximum timeout whether specified or not to prevent workflows from living forever.

### Context 

Similar to how the event-based start types have a `filters` field, event-consuming states have a `context` field. Like filters, the context field can restrict which events are considered matches by requiring an exact match on a CloudEvent context field.

Unlike filters, context values can be determined dynamically based on instance data. If a context value begins with "`{{`" and ends with "`}}`" everything inbetween will be evaluated as a `jq` command based on the instance data. 

## GenerateEvent State

Workflows can generate events for their namespace without relying on an Isolate using the GenerateEvent State. The fields for this state are fairly self-explanatory. Here's an example:

```yaml
- id: genEvent
  type: generateEvent
  event:
    type: "my.custom.event"
    source: "direktiv"
    data: '{ .input }'
    datacontenttype: "application/json"
```

If the `jq` command that populates the `data` field outputs a plain base64 encoded string and the `datacontenttype` field is set to anything other than `application/json` Direktiv will decode the string before sending the event.
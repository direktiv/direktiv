# Direktiv Workflow Definition

This document describes the rules for Direktiv workflow definition files. These files are written in YAML and dictate the behaviour of a workflow running on Direktiv. 

**Workflow**
```yaml
description: |
  A simple "Hello, world" demonstration.
states:
- id: hello
  type: noop
  transform: 'jq({ msg: "Hello, world!" })'
```

**Input**
```json
{}
```

**Output**
```json
{
	"msg": "Hello, world!"
}
```

Workflows have inputs and outputs, usually in JSON. Where examples appear in this document they will often be accompanied by inputs and outputs as seen above.

### WorkflowDefinition

This is the top-level structure of a Direktiv workflow definition. All workflows must have one.

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `url` | Link to further information. | string | no |
| `description` | Short description of the workflow.  | string | no |
| `functions` | List of function definitions for use by function-based `states`. | [[]FunctionDefinition](#FunctionDefinition) | no |
| `start` | Configuration for how the workflow should start. | [StartDefinition](#StartDefinition) | no |
| `states` | List of all possible workflow states. | [[]StateDefinition](#StateDefinition) | yes | 
| `timeouts` | Configuration of workflow-level timeouts. | [TimeoutsDefinition](#TimeoutsDefinition) | no |

## FunctionDefinition

Functions refer to anything executable by Direktiv as a unit of logic within a subflow that isn't otherwise part of basic state functionality. Usually this means either a purpose-built container or another workflow executed as a subflow. In some cases functions can be extensively configured, and they are often reused repeatedly within a workflow. To manage the size of Direktiv workflow definitions functions are predefined as much as possible and referenced when called.

These are the currently available function types:

* [`knative-global`](#GlobalKnativeFunctionDefinition)
* [`knative-namespace`](#NamespaceKnativeFunctionDefinition)
* [`knative-workflow`](#WorkflowKnativeFunctionDefinition)
* [`subflow`](#SubflowFunctionDefinition)
* [`kubernetes-job`](#KubernetesJobFunctionDefinition)

The following example demonstrate how to define and reference a function within a workflow:

**Workflow**
```yaml
description: |
  A basic demonstration of functions.
functions:
- type: knative-workflow
  id: request
  image: direktiv/request:latest
  size: small
states:
- id: getter
  type: action
  action:
    function: request
    input:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/todos/1"
```

**Input**
```json
{}
```

**Output**
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

### GlobalKnativeFunctionDefinition

A `knative-global` refers to a function that is implemented according to the [requirements](#TODO) for a direktiv knative service. Specifically, in this case referring to a service configured to be available "globally" (to all namespaces on the Direktiv servers).

This function type supports [`files`](#FunctionFileDefinition).

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [FunctionDefinition](#FunctionDefintion) is being used. In this case it must be set to `knative-global`. | string | yes | 
| `id` | A unique identifier for the function within the workflow definition. | string | yes |
| `service` | URI to a globally accessible function on the Direktiv servers. | string | yes |

### NamespacedKnativeFunctionDefinition

A `knative-namespace` refers to a function that is implemented according to the [requirements](#TODO) for a direktiv knative service. Specifically, in this case referring to a service configured to be available on the namespace.

This function type supports [`files`](#FunctionFileDefinition).

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [FunctionDefinition](#FunctionDefintion) is being used. In this case it must be set to `knative-namespace`. | string | yes | 
| `id` | A unique identifier for the function within the workflow definition. | string | yes |
| `service` | URI to a function on the namespace. | string | yes |

### WorkflowKnativeFunctionDefinition

A `knative-workflow` refers to a function that is implemented according to the [requirements](#TODO) for a direktiv knative service. Specifically, in this case referring to a service that Direktiv can create on-demand for the exclusive use by this workflow.

> Historically this was called `reusable`, but this keyword has been deprecated.

This function type supports [`files`](#FunctionFileDefinition).

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [FunctionDefinition](#FunctionDefintion) is being used. In this case it must be set to `knative-workflow`. | string | yes | 
| `id` | A unique identifier for the function within the workflow definition. | string | yes |
| `image` | URI to a `knative-workflow` compliant container. | string | yes |
| `size` | Specifies the container size. | [ContainerSizeDefinition](#ContainerSizeDefinition) | no |
| `cmd` | Custom command to execute within the container. | string | no |
| `scale` | Used as a suggestion to Direktiv for a minimum number of pods to keep running. Direktiv is not required to adhere to this minimum. The default value is zero, which may result in higher latency if a service goes unused for a while, but saves on resources. | integer | no |

#### ContainerSizeDefinition

When functions use containers you may be able to specify what size the container should be. This is done using one of three keywords, each representing a different [size preset](#TODO):

* `small`
* `medium`
* `large`

### SubflowFunctionDefinition

A `subflow` refers to a function that is actually another workflow. The other workflow is called with some input and its output is returned to this workflow.

This function type does not support [`files`](#FunctionFileDefinition).

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [FunctionDefinition](#FunctionDefintion) is being used. In this case it must be set to `subflow`. | string | yes | 
| `id` | A unique identifier for the function within the workflow definition. | string | yes |
| `workflow` | URI to a workflow within the same namespace. | string | yes |

### KubernetesJobFunctionDefinition

A `kubernetes-job` refers to a function that uses a container created to operate according to the requirements outlined [here](#TODO). 

> Historically this was called `isolated`, but this keyword has been deprecated.

This function type supports [`files`](#FunctionFileDefinition).

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [FunctionDefinition](#FunctionDefintion) is being used. In this case it must be set to `kubernetes-job`. | string | yes | 
| `id` | A unique identifier for the function within the workflow definition. | string | yes |
| `image` | URI to a `kubernetes-job` compliant container. | string | yes |
| `size` | Specifies the container size. | [ContainerSizeDefinition](#ContainerSizeDefinition) | no |
| `cmd` | Custom command to execute within the container. | string | no |

## StartDefinition

A `StartDefinition` may be defined using one of the following, depending on the desired behaviour:

* [`default`](#DefaultStartDefinition)
* [`scheduled`](#ScheduledStartDefinition)
* [`event`](#EventStartDefinition)
* [`eventsXor`](#EventsXorStartDefinition)
* [`eventsAnd`](#EventsAndStartDefinition)

If omitted from the workflow definition the [DefaultStartDefinition](#DefaultStartDefinition) will be used, which means the workflow will only be executed when called.

Regardless of which start definiton is used, all workflows can be called like a [DefaultStartDefinition](#DefaultStartDefinition). This is to make testing and debugging easier. To test properly the caller will need to simulate the start type's input data.

### DefaultStartDefinition

The default start definition is used for workflows that should only execute when called. This means subflows, workflows triggered by scripts, and workflows triggered manually by humans.

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StartDefinition](#StartDefintion) is being used. In this case it must be set to `default`. | string | yes | 
| `state` | References a defined state's `id`. This state will be used as the entrypoint into the workflow. If left undefined, it defaults to the first state defined in the `states` list.  | string | no |

### ScheduledStartDefinition

The scheduled start definition is used for workflows that should execute at regularly defined times. 

Scheduled workflows never have input data, so accurate testing should use `{}` as input. 

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StartDefinition](#StartDefintion) is being used. In this case it must be set to `scheduled`. | string | yes | 
| `state` | References a defined state's `id`. This state will be used as the entrypoint into the workflow. If left undefined, it defaults to the first state defined in the `states` list.  | string | no |
| `cron` | Defines the time(s) when the workflow should execute using a CRON expression. | string | yes |

**Example** (snippet)
```yaml
start:
  type: scheduled
  cron: '* * * * *' # Trigger a new instance every minute.
```

### EventStartDefinition 

The event start definition is used for workflows that should be executed whenever a relevant CloudEvents event is received. 

See [StartEventDefinition](#StartEventDefinition) for an explanation of the input data of event-triggered workflows.

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StartDefinition](#StartDefintion) is being used. In this case it must be set to `event`. | string | yes | 
| `state` | References a defined state's `id`. This state will be used as the entrypoint into the workflow. If left undefined, it defaults to the first state defined in the `states` list.  | string | no |
| `event` | Defines what events can trigger the workflow. | [StartEventDefinition](#StartEventDefinition) | yes |

### EventsXorStartDefinition 

The event "xor" start definition is used for workflows that should be executed whenever one of multiple possible CloudEvents events is received. 

See [StartEventDefinition](#StartEventDefinition) for an explanation of the input data of event-triggered workflows.

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StartDefinition](#StartDefintion) is being used. In this case it must be set to `eventsXor`. | string | yes | 
| `state` | References a defined state's `id`. This state will be used as the entrypoint into the workflow. If left undefined, it defaults to the first state defined in the `states` list.  | string | no |
| `events` | Defines what events can trigger the workflow.  | [[]StartEventDefinition](#StartEventDefinition) | yes |

### EventsAndStartDefinition 

The event "and" start definition is used for workflows that should be executed when multiple matching CloudEvents events are received. 

See [StartEventDefinition](#StartEventDefinition) for an explanation of the input data of event-triggered workflows.

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StartDefinition](#StartDefintion) is being used. In this case it must be set to `eventsAnd`. | string | yes | 
| `state` | References a defined state's `id`. This state will be used as the entrypoint into the workflow. If left undefined, it defaults to the first state defined in the `states` list.  | string | no |
| `lifespan` | An ISO8601 duration string. Sets the maximum duration an event can be stored before being discarded while waiting for other events. | string | no |
| `correlate` | CloudEvents event context keys can must exist on every event and have matching values to be grouped together. | []string | no |
| `events` | Defines what events can trigger the workflow.  | [[]StartEventDefinition](#StartEventDefinition) | yes |

### StartEventDefinition

The StartEventDefinition is a structure shared by various start definitions involving events. 

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which CloudEvents events can trigger the workflow by requiring an exact match to the event's own `type` context value. | string | yes | 
| `filters` | Optional key-value pairs to further restrict what events can trigger the workflow. For each pair, incoming CloudEvents context values will be checked for a match. All pairs must find a match for the event to be accepted. The "keys" are strings that match exactly to specific context keys, but the "values" can be "glob" patterns allowing them to match a range of possible context values. | object | no |

The input data of an event-triggered workflow is a JSON representation of all the received events stored under keys matching the events' respective type. For example, this CloudEvents event will result in the following input data in a workflow triggered by a single event:

**CloudEvents Event**
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

**Input Data**
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

## StateDefinition

A `StateDefinition` may be defined using one of the following, depending on the desired behaviour:

* [`action`](#ActionStateDefinition)
* [`consumeEvent`](#ConsumeEventStateDefinition)
* [`delay`](#DelayStateDefinition)
* [`error`](#ErrorStateDefinition)
* [`eventsAnd`](#EventsAndStateDefinition)
* [`eventsXor`](#EventsXorStateDefinition)
* [`foreach`](#ForeachStateDefinition)
* [`generateEvent`](#GenerateEventStateDefinition)
* [`getter`](#GetterStateDefinition)
* [`noop`](#NoopStateDefinition)
* [`parallel`](#ParallelStateDefinition)
* [`setter`](#SetterStateDefinition)
* [`switch`](#SwitchStateDefinition)
* [`validate`](#ValidateStateDefinition)

Many fields and concepts appear across multiple states. These topics are covered in more depth within their own sections:

* [Structured JQ](#StructuredJQ)
* [Error Handling](#StateErrorCatchers)
* [Logging](#StateLogging)
* [Metadata](#InstanceMetadata)
* [Transforms](#StateTransforms)

### ActionStateDefinition 

The `action` state is the simplest and most common way to call a function or invoke a workflow to act as a subflow. See [Actions](#TODO). 

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StateDefinition](#StateDefinition) is being used. In this case it must be set to `action`. | string | yes | 
| `id` | An identifier unique within the workflow to this one state. | string | yes |
| `log` | If defined, the workflow will generate a log when it commences this state. See [StateLogging](#StateLogging). | [Structured JQ](#StructuredJQ) | no |
| `metadata` | If defined, updates the instance's metadata. See [InstanceMetadata](#InstanceMetadata). | [Structured JQ](#StructuredJQ) | no |
| `transform` | If defined, modifies the instance's data upon completing the state logic. See [StateTransforms](#StateTransforms). | [Structured JQ](#StructuredJQ) | no |
| `transition` | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, this state terminates the workflow. | string | no |
| `catch` | Defines behaviour for handling of catchable errors. See [StateErrorCatchers](#StateErrorCatchers). | [[]ErrorCatchDefinition](#ErrorCatchDefinition) | no |
| `timeout` | ISO8601 duration string to set a non-default timeout. | string | no | 
| `aync` | If set to `true`, the workflow execution will continue without waiting for the action to return.  | boolean | no | 
| `action` | Defines the action to perform. | [ActionDefinition](#ActionDefinition) | yes |

### ActionDefinition 

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `function` | Name of the referenced function. See [FunctionDefinition](#FunctionDefinition). | string | yes |
| `input` | Selects or generates the data to send as input to the function. | [Structured JQ](#StructuredJQ) | no |
| `secrets` | Defines a list of secrets to temporarily add to the instance data under `.secrets`, before evaluating the `input`. | []string | no |
| `retries` | | [[]RetryPolicyDefinition](#RetryPolicyDefinition) | no |
| `files` | Determines a list of files to load onto the function's file-system from variables. Only valid if the referenced function supports it. | [[]FunctionFileDefinition](#FunctionFileDefinition) | no |

### FunctionFileDefinition

Some function types support loading variable directly from storage onto their file-systems. This object defines what variable to load and what to save it as.

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `key` | Identifies which variable to load into a file. | string | yes | 
| `scope` | Specifies the scope from which to load the variable. | [VariableScopeDefinition](#VariableScopeDefinition) | no |
| `as` | Names the resulting file. If left unspecified, the `key` will be used instead. | string | no |

### VariableScopeDefinition

Every variable exists within a single scope. The scope dictates what can access it and how persistent it is. There are three defined [scopes](#Variables):

* `instance`
* `workflow`
* `namespace`

### RetryPolicyDefinition 

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| codes | A list of "glob" patterns that will be compared to catchable error codes returned by the function to determine if this retry policy applies. | []string | yes |
| max_attempts | Maximum number of retry attempts. If the function has been retried this many times or more when this policy is invoked the retry will be skipped, and instead the error will be escalated to the state's error handling logic. See [StateErrorCatchers](#StateErrorCatchers) | integer | yes |
| delay | ISO8601 duration string giving a time delay between retry attempts. | string | no |
| multiplier | Value by which the delay is multiplied after each attempt. | float | no |

### ConsumeEventStateDefinition

To pause the workflow and wait until a CloudEvents event is received before proceeding, the `consumeEvent` is the simplest state that can be used. It is one of three states that can do so, along with `eventsAnd` and `eventsXor`.

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StateDefinition](#StateDefinition) is being used. In this case it must be set to `consumeEvent`. | string | yes | 
| `id` | An identifier unique within the workflow to this one state. | string | yes |
| `log` | If defined, the workflow will generate a log when it commences this state. See [StateLogging](#StateLogging). | [Structured JQ](#StructuredJQ) | no |
| `metadata` | If defined, updates the instance's metadata. See [InstanceMetadata](#InstanceMetadata). | [Structured JQ](#StructuredJQ) | no |
| `transform` | If defined, modifies the instance's data upon completing the state logic. See [StateTransforms](#StateTransforms). | [Structured JQ](#StructuredJQ) | no |
| `transition` | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, this state terminates the workflow. | string | no |
| `catch` | Defines behaviour for handling of catchable errors. See [StateErrorCatchers](#StateErrorCatchers). | [[]ErrorCatchDefinition](#ErrorCatchDefinition) | no |
| `timeout` | An ISO8601 duration string. | string | no |
| `event` | Defines the criteria by which incoming CloudEvents events are evaluated to find a match. | [ConsumeEventDefinition](#ConsumeEventDefinition) | yes |

### ConsumeEventDefinition

The StartEventDefinition is a structure shared by various event-consuming states. 

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which CloudEvents events can trigger the workflow by requiring an exact match to the event's own `type` context value. | string | yes | 
| `context` | Optional key-value pairs to further restrict what events can trigger the workflow. For each pair, incoming CloudEvents context values will be checked for a match. All pairs must find a match for the event to be accepted. The "keys" must be strings that match exactly to specific context keys, but the "values" can be "glob" patterns allowing them to match a range of possible context values. | [Structured JQ](#StructuredJQ) | no |

The received data of an event-triggered workflow is a JSON representation of all the received events stored under keys matching the events' respective type. For example, this CloudEvents event will result in the following data:

**CloudEvents Event**
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

**Input Data**
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

### DelayStateDefinition

If the workflow needs to pause for a specific length of time, the delay state is usually the simplest way to do that.

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StateDefinition](#StateDefinition) is being used. In this case it must be set to `delay`. | string | yes | 
| `id` | An identifier unique within the workflow to this one state. | string | yes |
| `log` | If defined, the workflow will generate a log when it commences this state. See [StateLogging](#StateLogging). | [Structured JQ](#StructuredJQ) | no |
| `metadata` | If defined, updates the instance's metadata. See [InstanceMetadata](#InstanceMetadata). | [Structured JQ](#StructuredJQ) | no |
| `transform` | If defined, modifies the instance's data upon completing the state logic. See [StateTransforms](#StateTransforms). | [Structured JQ](#StructuredJQ) | no |
| `transition` | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, this state terminates the workflow. | string | no |
| `catch` | Defines behaviour for handling of catchable errors. See [StateErrorCatchers](#StateErrorCatchers). | [[]ErrorCatchDefinition](#ErrorCatchDefinition) | no |
| `duration` | An ISO8601 duration string. | string | yes |

### ErrorStateDefinition 

When workflow logic end up in a failure mode, the `error` state can be used to mark the instance as failed. This allows the instance to report what went wrong to the caller, which may then be handled or reported appropriately.

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StateDefinition](#StateDefinition) is being used. In this case it must be set to `error`. | string | yes | 
| `id` | An identifier unique within the workflow to this one state. | string | yes |
| `log` | If defined, the workflow will generate a log when it commences this state. See [StateLogging](#StateLogging). | [Structured JQ](#StructuredJQ) | no |
| `metadata` | If defined, updates the instance's metadata. See [InstanceMetadata](#InstanceMetadata). | [Structured JQ](#StructuredJQ) | no |
| `transform` | If defined, modifies the instance's data upon completing the state logic. See [StateTransforms](#StateTransforms). | [Structured JQ](#StructuredJQ) | no |
| `transition` | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, this state terminates the workflow. | string | no |
| `catch` | Defines behaviour for handling of catchable errors. See [StateErrorCatchers](#StateErrorCatchers). | [[]ErrorCatchDefinition](#ErrorCatchDefinition) | no |
| `error` | A short descriptive error code that can be caught by a parent workflow. | string | yes |
| `message` | Generates a more detailed message or object that can contain instance data, to provide more information for users. | [Structured JQ](#StructuredJQ) | yes |

### EventsAndStateDefinition 

To pause the workflow and wait until multiple CloudEvents events are received before proceeding, the `eventsAnd` is used. Every listed event must be received for the state to complete.

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StateDefinition](#StateDefinition) is being used. In this case it must be set to `eventsAnd`. | string | yes | 
| `id` | An identifier unique within the workflow to this one state. | string | yes |
| `log` | If defined, the workflow will generate a log when it commences this state. See [StateLogging](#StateLogging). | [Structured JQ](#StructuredJQ) | no |
| `metadata` | If defined, updates the instance's metadata. See [InstanceMetadata](#InstanceMetadata). | [Structured JQ](#StructuredJQ) | no |
| `transform` | If defined, modifies the instance's data upon completing the state logic. See [StateTransforms](#StateTransforms). | [Structured JQ](#StructuredJQ) | no |
| `transition` | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, this state terminates the workflow. | string | no |
| `catch` | Defines behaviour for handling of catchable errors. See [StateErrorCatchers](#StateErrorCatchers). | [[]ErrorCatchDefinition](#ErrorCatchDefinition) | no |
| `timeout` | An ISO8601 duration string. | string | no |
| `events` | Defines the criteria by which incoming CloudEvents events are evaluated to find a match. | [ConsumeEventDefinition](#ConsumeEventDefinition) | yes |

### EventsXorStateDefinition 

To pause the workflow and wait until one of multiple CloudEvents events is received before proceeding, the `eventsXor` state might be used. Any event match received will cause this state to complete.

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StateDefinition](#StateDefinition) is being used. In this case it must be set to `eventsXor`. | string | yes | 
| `id` | An identifier unique within the workflow to this one state. | string | yes |
| `log` | If defined, the workflow will generate a log when it commences this state. See [StateLogging](#StateLogging). | [Structured JQ](#StructuredJQ) | no |
| `metadata` | If defined, updates the instance's metadata. See [InstanceMetadata](#InstanceMetadata). | [Structured JQ](#StructuredJQ) | no |
| `transform` | If defined, modifies the instance's data upon completing the state logic. See [StateTransforms](#StateTransforms). | [Structured JQ](#StructuredJQ) | no |
| `transition` | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, this state terminates the workflow. | string | no |
| `catch` | Defines behaviour for handling of catchable errors. See [StateErrorCatchers](#StateErrorCatchers). | [[]ErrorCatchDefinition](#ErrorCatchDefinition) | no |
| `timeout` | An ISO8601 duration string. | string | no |
| `events` | Defines the criteria by which incoming CloudEvents events are evaluated to find a match. | [ConsumeEventDefinition](#ConsumeEventDefinition) | yes |

### ForeachStateDefinition

The `foreach` state is a convenient way to divide some data and then perform an action on each element in parallel. 

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StateDefinition](#StateDefinition) is being used. In this case it must be set to `foreach`. | string | yes | 
| `id` | An identifier unique within the workflow to this one state. | string | yes |
| `log` | If defined, the workflow will generate a log when it commences this state. See [StateLogging](#StateLogging). | [Structured JQ](#StructuredJQ) | no |
| `metadata` | If defined, updates the instance's metadata. See [InstanceMetadata](#InstanceMetadata). | [Structured JQ](#StructuredJQ) | no |
| `transform` | If defined, modifies the instance's data upon completing the state logic. See [StateTransforms](#StateTransforms). | [Structured JQ](#StructuredJQ) | no |
| `transition` | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, this state terminates the workflow. | string | no |
| `catch` | Defines behaviour for handling of catchable errors. See [StateErrorCatchers](#StateErrorCatchers). | [[]ErrorCatchDefinition](#ErrorCatchDefinition) | no |
| `timeout` | ISO8601 duration string to set a non-default timeout. | string | no | 
| `array` | Selects or generates an array, from which each element will be separately acted upon. The `action.input` will be evaluated against each element in this array, rather than the usual instance data. | [Structured JQ](#StructuredJQ) | yes | 
| `action` | Defines the action to perform. | [ActionDefinition](#ActionDefinition) | yes |

### GenerateEventStateDefinition 

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StateDefinition](#StateDefinition) is being used. In this case it must be set to `generateEvent`. | string | yes | 
| `id` | An identifier unique within the workflow to this one state. | string | yes |
| `log` | If defined, the workflow will generate a log when it commences this state. See [StateLogging](#StateLogging). | [Structured JQ](#StructuredJQ) | no |
| `metadata` | If defined, updates the instance's metadata. See [InstanceMetadata](#InstanceMetadata). | [Structured JQ](#StructuredJQ) | no |
| `transform` | If defined, modifies the instance's data upon completing the state logic. See [StateTransforms](#StateTransforms). | [Structured JQ](#StructuredJQ) | no |
| `transition` | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, this state terminates the workflow. | string | no |
| `catch` | Defines behaviour for handling of catchable errors. See [StateErrorCatchers](#StateErrorCatchers). | [[]ErrorCatchDefinition](#ErrorCatchDefinition) | no |
| `delay` | ISO8601 duration string defining how long to hold the event before broadcasting it. | string | no |
| `event` | Defines the event to generate. | [GenerateEventDefinition](#GenerateEventDefinition) | yes |

### GenerateEventDefinition

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Sets the CloudEvents event type. | string | yes |
| `source` | Sets the CloudEvents event source. | string | yes |
| `data` | Defines the content of the payload for the CloudEvents event. | [Structured JQ](#StructuredJQ) | no |
| `datacontenttype` | An RFC2046 string specifying the payload content type. | string | no |
| `context` | If defined, must evaluate to an object of key-value pairs. These will be used to define CloudEvents event context data. | [Structured JQ](#StructuredJQ) | no |

### GetterStateDefinition 

To load variables, use the `getter` state. See [Variables](#Variables).

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StateDefinition](#StateDefinition) is being used. In this case it must be set to `getter`. | string | yes | 
| `id` | An identifier unique within the workflow to this one state. | string | yes |
| `log` | If defined, the workflow will generate a log when it commences this state. See [StateLogging](#StateLogging). | [Structured JQ](#StructuredJQ) | no |
| `metadata` | If defined, updates the instance's metadata. See [InstanceMetadata](#InstanceMetadata). | [Structured JQ](#StructuredJQ) | no |
| `transform` | If defined, modifies the instance's data upon completing the state logic. See [StateTransforms](#StateTransforms). | [Structured JQ](#StructuredJQ) | no |
| `transition` | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, this state terminates the workflow. | string | no |
| `catch` | Defines behaviour for handling of catchable errors. See [StateErrorCatchers](#StateErrorCatchers). | [[]ErrorCatchDefinition](#ErrorCatchDefinition) | no |
| `variables` | Defines variables to load. | [[]VariableSetterDefinition](#VariableGetterDefinition) | yes |

### VariableGetterDefinition

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `key` | Variable name. | string | yes |
| `scope` | Selects the scope to which the variable belongs. If undefined, defaults to `instance`. See [Variables](#Variables). | yes | no |
| `as` | Names the resulting data. If left unspecified, the `key` will be used instead. | string | no |

### NoopStateDefinition

Often workflows need to do something that can be achieved using logic built into most state types. For example, to log something, or to transform the instance data by running a `jq` command. In many cases this can be done by an existing state within the workflow, but sometimes it's necessary to split it out into a separate state. The `noop` state exists for this purpose.

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StateDefinition](#StateDefinition) is being used. In this case it must be set to `noop`. | string | yes | 
| `id` | An identifier unique within the workflow to this one state. | string | yes |
| `log` | If defined, the workflow will generate a log when it commences this state. See [StateLogging](#StateLogging). | [Structured JQ](#StructuredJQ) | no |
| `metadata` | If defined, updates the instance's metadata. See [InstanceMetadata](#InstanceMetadata). | [Structured JQ](#StructuredJQ) | no |
| `transform` | If defined, modifies the instance's data upon completing the state logic. See [StateTransforms](#StateTransforms). | [Structured JQ](#StructuredJQ) | no |
| `transition` | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, this state terminates the workflow. | string | no |
| `catch` | Defines behaviour for handling of catchable errors. See [StateErrorCatchers](#StateErrorCatchers). | [[]ErrorCatchDefinition](#ErrorCatchDefinition) | no |

### ParallelStateDefinition 

The `parallel` state is an alternative to the `action` state when a workflow can perform multiple threads of logic simultaneously. 

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StateDefinition](#StateDefinition) is being used. In this case it must be set to `parallel`. | string | yes | 
| `id` | An identifier unique within the workflow to this one state. | string | yes |
| `log` | If defined, the workflow will generate a log when it commences this state. See [StateLogging](#StateLogging). | [Structured JQ](#StructuredJQ) | no |
| `metadata` | If defined, updates the instance's metadata. See [InstanceMetadata](#InstanceMetadata). | [Structured JQ](#StructuredJQ) | no |
| `transform` | If defined, modifies the instance's data upon completing the state logic. See [StateTransforms](#StateTransforms). | [Structured JQ](#StructuredJQ) | no |
| `transition` | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, this state terminates the workflow. | string | no |
| `catch` | Defines behaviour for handling of catchable errors. See [StateErrorCatchers](#StateErrorCatchers). | [[]ErrorCatchDefinition](#ErrorCatchDefinition) | no |
| `timeout` | ISO8601 duration string to set a non-default timeout. | string | no | 
| `mode` | If defined, must be either `and` or `or`. The default is `and`. This setting determines whether the state is considered successfully completed only if all threads have returned without error (`and`) or as soon as any single thread returns without error (`or`). | string | no | 
| `actions` | Defines the action to perform. | [[]ActionDefinition](#ActionDefinition) | yes |

### SetterStateDefinition 

To create or change variables, use the `setter` state. See [Variables](#Variables).

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StateDefinition](#StateDefinition) is being used. In this case it must be set to `setter`. | string | yes | 
| `id` | An identifier unique within the workflow to this one state. | string | yes |
| `log` | If defined, the workflow will generate a log when it commences this state. See [StateLogging](#StateLogging). | [Structured JQ](#StructuredJQ) | no |
| `metadata` | If defined, updates the instance's metadata. See [InstanceMetadata](#InstanceMetadata). | [Structured JQ](#StructuredJQ) | no |
| `transform` | If defined, modifies the instance's data upon completing the state logic. See [StateTransforms](#StateTransforms). | [Structured JQ](#StructuredJQ) | no |
| `transition` | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, this state terminates the workflow. | string | no |
| `catch` | Defines behaviour for handling of catchable errors. See [StateErrorCatchers](#StateErrorCatchers). | [[]ErrorCatchDefinition](#ErrorCatchDefinition) | no |
| `variables` | Defines variables to push. | [[]VariableSetterDefinition](#VariableSetterDefinition) | yes |

### VariableSetterDefinition

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `key` | Variable name. | string | yes |
| `scope` | Selects the scope to which the variable belongs. If undefined, defaults to `instance`. See [Variables](#Variables). | yes | no |
| `mimeType` | Store a MIME type with the variable. If left undefined, it will default to `application/json`. Two specific MIME types cause this state to behave differently: `text/plain` and `application/octet-stream`. If the `value` evaluates to a JSON string the MIME type is `text/plain`, that string will be stored in plaintext (without JSON quotes and escapes). If if the `value` is a JSON string containing base64 encoded data and the MIME type is `application/octet-stream`, the base64 data will be decoded and stored as binary data. | string | no |
| `value` | Select or generate the data to store.  | [Structured JQ](#StructuredJQ) | yes |

### SwitchStateDefinition

To change the behaviour of a workflow based on the instance data, use a `switch` state. This state does nothing except choose between any number of different possible state transitions.

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StateDefinition](#StateDefinition) is being used. In this case it must be set to `switch`. | string | yes | 
| `id` | An identifier unique within the workflow to this one state. | string | yes |
| `log` | If defined, the workflow will generate a log when it commences this state. See [StateLogging](#StateLogging). | [Structured JQ](#StructuredJQ) | no |
| `metadata` | If defined, updates the instance's metadata. See [InstanceMetadata](#InstanceMetadata). | [Structured JQ](#StructuredJQ) | no |
| `defaultTransform` | If defined, modifies the instance's data upon completing the state logic. But only if none of the `conditions` are met. See [StateTransforms](#StateTransforms). | [Structured JQ](#StructuredJQ) | no |
| `defaultTransition` | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, this state terminates the workflow. But only if none of the `conditions` are met. | string | no |
| `catch` | Defines behaviour for handling of catchable errors. See [StateErrorCatchers](#StateErrorCatchers). | [[]ErrorCatchDefinition](#ErrorCatchDefinition) | no |
| `conditions` | List of conditions, which are evaluated in-order until a match is found. | [[]SwitchConditionDefinition](#SwitchConditionDefinition) | yes |

### SwitchConditionDefinition 

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `condition` | Selects or generates the data used to determine if condition is met. The condition is considered met if the result is anything other than `null`, `false`, `{}`, `[]`, `""`, or `0`. | [Structured JQ](#StructuredJQ) | yes | 
| `transform` | If defined, modifies the instance's data upon completing the state logic. See [StateTransforms](#StateTransforms). | [Structured JQ](#StructuredJQ) | no | 
| `transition` | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, matching this condition terminates the workflow. | string | no | 

### ValidateStateDefinition

Since workflows receive external input it may be necessary to check that instance data is valid. The `validate` state exists for this purpose. 

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `type` | Identifies which kind of [StateDefinition](#StateDefinition) is being used. In this case it must be set to `validate`. | string | yes | 
| `id` | An identifier unique within the workflow to this one state. | string | yes |
| `log` | If defined, the workflow will generate a log when it commences this state. See [StateLogging](#StateLogging). | [Structured JQ](#StructuredJQ) | no |
| `metadata` | If defined, updates the instance's metadata. See [InstanceMetadata](#InstanceMetadata). | [Structured JQ](#StructuredJQ) | no |
| `transform` | If defined, modifies the instance's data upon completing the state logic. See [StateTransforms](#StateTransforms). | [Structured JQ](#StructuredJQ) | no |
| `transition` | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, this state terminates the workflow. | string | no |
| `catch` | Defines behaviour for handling of catchable errors. See [StateErrorCatchers](#StateErrorCatchers). | [[]ErrorCatchDefinition](#ErrorCatchDefinition) | no |
| `subject` | Selects or generates the data that will be compared to the `schema`. If undefined, it will be default to `'jq(.)'`. | [Structured JQ](#StructuredJQ) | no |
| `schema` | A YAMLified representation of a JSON Schema that defines whether the `subject` is considered valid. | object | yes |

### TimeoutsDefinition

In addition to any timeouts applied on a state-by-state basis, every workflow has two global timeouts that begin ticking from the moment the workflow starts. This is where you can configure these timeouts differently to their defaults.

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `interrupt` | An ISO8601 duration string. Sets the time to wait before throwing a catchable `direktiv.cancels.timeout.soft` error. Consider this a soft timeout. | string | no |
| `kill` | An ISO8601 duration string. Sets the time to wait before throwing an uncatchable `direktiv.cancels.timeout.hard` error. This is a hard timeout. | string | no |

### StructuredJQ

To make Direktiv powerful and flexible we use `jq` in many places. Wherever a type called Structured JQ appears you can input generic data in any form you like, and it will be converted from YAML to JSON. For example, this log field will output the following JSON:

**YAML**
```yaml
log: 
  a: 5
```

**JSON**
```json
{
  "a": 5
}
```

Structured JQ also inspects all strings in the YAML searching for embedded queries, executing them and inserting the results into the data. Embeddeed queries are wrapped within the brackets of `jq()`, as in the following example:

**Instance Data**
```json
{
  "x": 5
}
```

**YAML**
```yaml
log: 
  a: 'jq(.x)'
```

**JSON**
```json
{
  "a": 5
}
```

If the JQ brackets the entire string the entire field will be replaced with the results of the query, as in the above example where `a` is represented as a JSON integer. Otherwise the results will be marshalled into JSON and be inserted as a string into the parent string. 

**Instance Data**
```json
{
  "x": 5
}
```

**YAML**
```yaml
log: 
  a: 'x = jq(.x)'
```

**JSON**
```json
{
  "a": "x = 5"
}
```

These JQ substitutions are not limited to primitive types. Entire objects or arrays, or anything else is all supported. 

**Instance Data**
```json
{
  "x": 5
}
```

**YAML**
```yaml
log: 'jq(.)'
```

**JSON**
```json
{
  "x": 5
}
```

### StateErrorCatchers

TODO

### StateLogging

The most common type to use is a string, which may contain structured `jq`, but any types are valid since the generated logs are JSON.

TODO 

### InstanceMetadata 

The workflow will replace the instance's metadata with the provided value. Can be any type, and supports structured `jq`.

TODO 

### StateTransforms

The workflow will modify its memory just before it completes the state. The evaluated result of the transform must be an object.

TODO 

### Variables

Most workflows can get by with their instance data, but in some cases that may be insufficient. Instance data has a limited maximum size, and on its own it cannot be used to persist data long-term. Variables are stored on Direktiv as a means of storing and retrieving data between functions, instances, or workflows. 

All variables belong to a `scope`. The scopes are `instance`, `workflow`, and `namespace`. Instance scoped variables are only accessible to the singular instance that created them. Workflow scoped variables can be used and shared between multiple instances of the same workflow. Namespace scoped variables are available to all instances of all workflows on the namespace. 

All variables are identified by a name, and each name is unique within its scope. The main ways for workflows to interact with variables is through [`getter`](#GetterStateDefinition) and [`setter`](#SetterStateDefinition) states, though other methods are also possible. 

In addition to the scopes outlined above, there is a special `system` scope. This scope is a utility to make miscellaneous information accessible to an instance. The following special variables exist in the system scope:

| Key | Description |
| --- | --- |
| `instance` | Returns the instance ID of the running instance. |
| `uuid` | Returns a randomly generated UUID. |
| `epoch` | Returns the current time in unix/epoch format. |

# TODO

* Add `url` field to workflow definition.
* What happened to Exclusive / Singular field?
* Check if states use `eventAnd` or `eventsAnd`. Consider changing for consistency?
* Links to container specs.
* Links to isolated specs.
* Rename `isolated` and `reusable` to be consistent with other function types.
* Link to size presets documentation.
* Explain function files.
* Function file type (plain/base64/tar/tar.gz) needs revisiting
* Explain secrets.

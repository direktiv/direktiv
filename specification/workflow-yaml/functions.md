# Functions 

## FunctionDefinition

Functions refer to anything executable by Direktiv as a unit of logic within a subflow that isn't otherwise part of basic state functionality. Usually this means either a purpose-built container or another workflow executed as a subflow. In some cases functions can be extensively configured, and they are often reused repeatedly within a workflow. To manage the size of Direktiv workflow definitions functions are predefined as much as possible and referenced when called.

These are the currently available function types:

- [Functions](#functions)
	- [FunctionDefinition](#functiondefinition)
		- [GlobalKnativeFunctionDefinition](#globalknativefunctiondefinition)
		- [NamespacedKnativeFunctionDefinition](#namespacedknativefunctiondefinition)
		- [WorkflowKnativeFunctionDefinition](#workflowknativefunctiondefinition)
			- [ContainerSizeDefinition](#containersizedefinition)
		- [SubflowFunctionDefinition](#subflowfunctiondefinition)

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

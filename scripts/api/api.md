


# Direktiv API.
Direktiv Open API Specification
Direktiv Documentation can be found at https://docs.direktiv.io/
  

## Informations

### Version

1.0.0

### Contact

 info@direktiv.io 

## Content negotiation

### URI Schemes
  * http
  * https

### Consumes
  * application/json
  * text/plain

### Produces
  * application/json
  * text/event-stream

## Access control

### Security Schemes

#### api_key (header: KEY)



> **Type**: apikey

### Security Requirements
  * api_key

## All endpoints

###  directory

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| PUT | /api/namespaces/{namespace}/tree/{directory}?op=create-directory | [create directory](#create-directory) | Create a Directory |
| GET | /api/namespaces/{namespace}/tree/{nodePath} | [get nodes](#get-nodes) | Get List of Namespace Nodes |
  


###  global_services

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| POST | /api/functions | [create global service](#create-global-service) | Create Global Service |
| DELETE | /api/functions/{serviceName}/revisions/{revisionGeneration} | [delete global revision](#delete-global-revision) | Delete Global Service Revision |
| DELETE | /api/functions/{serviceName} | [delete global service](#delete-global-service) | Delete Global Service |
| GET | /api/functions/{serviceName} | [get global service](#get-global-service) | Get Global Service Details |
| GET | /api/functions | [get global service list](#get-global-service-list) | Get Global Services List |
| GET | /api/functions/{serviceName}/revisions/{revisionGeneration}/pods | [list global service revision pods](#list-global-service-revision-pods) | Get Global Service Revision Pods List |
| POST | /api/functions/{serviceName} | [update global service](#update-global-service) | Create Global Service Revision |
| PATCH | /api/functions/{serviceName} | [update global service traffic](#update-global-service-traffic) | Update Global Service Traffic |
| GET | /api/functions/{serviceName}/revisions/{revisionGeneration} | [watch global service revision](#watch-global-service-revision) | Watch Global Service Revision |
| GET | /api/functions/{serviceName}/revisions | [watch global service revision list](#watch-global-service-revision-list) | Watch Global Service Revision List |
  


###  instances

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| POST | /api/namespaces/{namespace}/instances/{instance}/cancel | [cancel instance](#cancel-instance) | Cancel a Pending Instance |
| GET | /api/namespaces/{namespace}/instances/{instance} | [get instance](#get-instance) | Get a Instance |
| GET | /api/namespaces/{namespace}/instances/{instance}/input | [get instance input](#get-instance-input) | Get a Instance Input |
| GET | /api/namespaces/{namespace}/instances | [get instance list](#get-instance-list) | Get List Instances |
| GET | /api/namespaces/{namespace}/instances/{instance}/output | [get instance output](#get-instance-output) | Get a Instance Output |
  


###  logs

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/namespaces/{namespace}/tree/{workflow}?op=logs | [get workflow logs](#get-workflow-logs) | Get Workflow Level Logs |
| GET | /api/namespaces/{namespace}/instances/{instance}/logs | [instance logs](#instance-logs) | Gets Instance Logs |
| GET | /api/namespaces/{namespace}/logs | [namespace logs](#namespace-logs) | Gets Namespace Level Logs |
| GET | /api/logs | [server logs](#server-logs) | Get Direktiv Server Logs |
  


###  metrics

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/namespaces/{namespace}/metrics/failed | [namespace metrics failed](#namespace-metrics-failed) | Gets Namespace Failed Workflow Instances Metrics |
| GET | /api/namespaces/{namespace}/metrics/invoked | [namespace metrics invoked](#namespace-metrics-invoked) | Gets Namespace Invoked Workflow Metrics |
| GET | /api/namespaces/{namespace}/metrics/milliseconds | [namespace metrics milliseconds](#namespace-metrics-milliseconds) | Gets Namespace Workflow Timing Metrics |
| GET | /api/namespaces/{namespace}/metrics/successful | [namespace metrics successful](#namespace-metrics-successful) | Gets Namespace Successful Workflow Instances Metrics |
| GET | /api/namespaces/{namespace}/tree/{workflow}?op=metrics-invoked | [workflow metrics invoked](#workflow-metrics-invoked) | Gets Invoked Workflow Metrics |
| GET | /api/namespaces/{namespace}/tree/{workflow}?op=metrics-failed | [workflow metrics milliseconds](#workflow-metrics-milliseconds) | Gets Workflow Time Metrics |
| GET | /api/namespaces/{namespace}/tree/{workflow}?op=metrics-state-milliseconds | [workflow metrics state milliseconds](#workflow-metrics-state-milliseconds) | Gets a Workflow State Time Metrics |
| GET | /api/namespaces/{namespace}/tree/{workflow}?op=metrics-successful | [workflow metrics successful](#workflow-metrics-successful) | Gets Successful Workflow Metrics |
  


###  namespace_services

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| POST | /api/functions/namespaces/{namespace} | [create namespace service](#create-namespace-service) | Create Namespace Service |
| DELETE | /api/functions/namespaces/{namespace}/function/{serviceName}/revisions/{revisionGeneration} | [delete namespace revision](#delete-namespace-revision) | Delete Namespace Service Revision |
| DELETE | /api/functions/namespaces/{namespace}/function/{serviceName} | [delete namespace service](#delete-namespace-service) | Delete Namespace Service |
| GET | /api/functions/namespaces/{namespace}/function/{serviceName} | [get namespace service](#get-namespace-service) | Get Namespace Service Details |
| GET | /api/functions/namespaces/{namespace} | [get namespace service list](#get-namespace-service-list) | Get Namespace Services List |
| GET | /api/functions/namespaces/{namespace}/function/{serviceName}/revisions/{revisionGeneration}/pods | [list namespace service revision pods](#list-namespace-service-revision-pods) | Get Namespace Service Revision Pods List |
| POST | /api/functions/namespaces/{namespace}/function/{serviceName} | [update namespace service](#update-namespace-service) | Create Namespace Service Revision |
| PATCH | /api/functions/namespaces/{namespace}/function/{serviceName} | [update namespace service traffic](#update-namespace-service-traffic) | Update Namespace Service Traffic |
| GET | /api/functions/namespaces/{namespace}/function/{serviceName}/revisions/{revisionGeneration} | [watch namespace service revision](#watch-namespace-service-revision) | Watch Namespace Service Revision |
| GET | /api/functions/namespaces/{namespace}/function/{serviceName}/revisions | [watch namespace service revision list](#watch-namespace-service-revision-list) | Watch Namespace Service Revision List |
  


###  namespaces

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| PUT | /api/namespaces/{namespace} | [create namespace](#create-namespace) | Creates a namespace |
| DELETE | /api/namespaces/{namespace} | [delete namespace](#delete-namespace) | Delete a namespace |
| GET | /api/namespaces | [get namespaces](#get-namespaces) | Gets the list of namespaces |
  


###  other

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| POST | /api/namespaces/{namespace}/broadcast | [broadcast cloudevent](#broadcast-cloudevent) | Broadcast Cloud Event |
| POST | /api/jq | [jq playground](#jq-playground) | JQ Playground api to test jq queries |
  


###  registries

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| POST | /api/namespaces/{namespace}/registries | [delete registry](#delete-registry) | Delete a Namespace Container Registry |
| GET | /api/namespaces/{namespace}/registries | [get registries](#get-registries) | Get List of Namespace Registries |
  


###  secrets

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| PUT | /api/namespaces/{namespace}/secrets/{secret} | [create secret](#create-secret) | Create a Namespace Secret |
| DELETE | /api/namespaces/{namespace}/secrets/{secret} | [delete secret](#delete-secret) | Delete a Namespace Secret |
| GET | /api/namespaces/{namespace}/secrets | [get secrets](#get-secrets) | Get List of Namespace Secrets |
  


###  variables

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| DELETE | /api/namespaces/{namespace}/instances/{instance}/vars/{variable} | [delete instance variable](#delete-instance-variable) | Delete a Instance Variable |
| DELETE | /api/namespaces/{namespace}/vars/{variable} | [delete namespace variable](#delete-namespace-variable) | Delete a Namespace Variable |
| DELETE | /api/namespaces/{namespace}/tree/{workflow}?op=delete-var | [delete workflow variable](#delete-workflow-variable) | Delete a Workflow Variable |
| GET | /api/namespaces/{namespace}/instances/{instance}/vars/{variable} | [get instance variable](#get-instance-variable) | Get a Instance Variable |
| GET | /api/namespaces/{namespace}/instances/{instance}/vars | [get instance variables](#get-instance-variables) | Get List of Instance Variable |
| GET | /api/namespaces/{namespace}/vars/{variable} | [get namespace variable](#get-namespace-variable) | Get a Namespace Variable |
| GET | /api/namespaces/{namespace}/vars | [get namespace variables](#get-namespace-variables) | Get Namespace Variable List |
| GET | /api/namespaces/{namespace}/tree/{workflow}?op=var | [get workflow variable](#get-workflow-variable) | Get a Workflow Variable |
| GET | /api/namespaces/{namespace}/tree/{workflow}?op=vars | [get workflow variables](#get-workflow-variables) | Get List of Workflow Variables |
| PUT | /api/namespaces/{namespace}/instances/{instance}/vars/{variable} | [set instance variable](#set-instance-variable) | Set a Instance Variable |
| PUT | /api/namespaces/{namespace}/vars/{variable} | [set namespace variable](#set-namespace-variable) | Set a Namespace Variable |
| PUT | /api/namespaces/{namespace}/tree/{workflow}?op=set-var | [set workflow variable](#set-workflow-variable) | Set a Workflow Variable |
  


###  workflow_services

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/functions/namespaces/{namespace}/tree/{workflow}?op=function | [get workflow service](#get-workflow-service) | Get Workflow Service Details |
| GET | /api/functions/namespaces/{namespace}/tree/{workflow}?op=function-revision | [get workflow service revision](#get-workflow-service-revision) | Get Workflow Service Revision |
| GET | /api/functions/namespaces/{namespace}/tree/{workflow}?op=function-revisions | [get workflow service revision list](#get-workflow-service-revision-list) | Get Workflow Service Revision List |
| GET | /api/functions/namespaces/{namespace}/tree/{workflow}?op=pods | [list workflow service revision pods](#list-workflow-service-revision-pods) | Get Workflow Service Revision Pods List |
| GET | /api/functions/namespaces/{namespace}/tree/{workflow}?op=services | [list workflow services](#list-workflow-services) | Get Workflow Services List |
  


###  workflows

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/namespaces/{namespace}/tree/{workflow}?op=wait | [await execute workflow](#await-execute-workflow) | Await Execute a Workflow |
| POST | /api/namespaces/{namespace}/tree/{workflow}?op=wait | [await execute workflow body](#await-execute-workflow-body) | Await Execute a Workflow With Body |
| PUT | /api/namespaces/{namespace}/tree/{workflow}?op=create-workflow | [create workflow](#create-workflow) | Create a Workflow |
| POST | /api/namespaces/{namespace}/tree/{workflow}?op=execute | [execute workflow](#execute-workflow) | Execute a Workflow |
| POST | /api/namespaces/{namespace}/tree/{workflow}?op=set-workflow-event-logging | [set workflow cloud event logs](#set-workflow-cloud-event-logs) | Set Cloud Event for Workflow to Log to |
| POST | /api/namespaces/{namespace}/tree/{workflow}?op=toggle | [toggle workflow](#toggle-workflow) | Set Cloud Event for Workflow to Log to |
| POST | /api/namespaces/{namespace}/tree/{workflow}?op=update-workflow | [update workflow](#update-workflow) | Update a Workflow |
  


## Paths

### <span id="await-execute-workflow"></span> Await Execute a Workflow (*awaitExecuteWorkflow*)

```
GET /api/namespaces/{namespace}/tree/{workflow}?op=wait
```

Executes a workflow. This path will wait until the workflow execution has completed and return the instance output.
NOTE: Input can also be provided with the `input.X` query parameters; Where `X` is the json key.
Only top level json keys are supported when providing input with query parameters.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| workflow | `path` | string | `string` |  | ✓ |  | path to target workflow |
| ctype | `query` | string | `string` |  |  |  | Manually set the Content-Type response header instead of auto-detected. This doesn't change the body of the response in any way. |
| field | `query` | string | `string` |  |  |  | If provided, instead of returning the entire output json the response body will contain the single top-level json field |
| raw-output | `query` | boolean | `bool` |  |  |  | If set to true, will return an empty output as null, encoded base64 data as decoded binary data, and quoted json strings as a escaped string. |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#await-execute-workflow-200) | OK | successfully executed workflow |  | [schema](#await-execute-workflow-200-schema) |

#### Responses


##### <span id="await-execute-workflow-200"></span> 200 - successfully executed workflow
Status: OK

###### <span id="await-execute-workflow-200-schema"></span> Schema

### <span id="await-execute-workflow-body"></span> Await Execute a Workflow With Body (*awaitExecuteWorkflowBody*)

```
POST /api/namespaces/{namespace}/tree/{workflow}?op=wait
```

Executes a workflow with optionally some input provided in the request body as json.
This path will wait until the workflow execution has completed and return the instance output.
NOTE: Input can also be provided with the `input.X` query parameters; Where `X` is the json key.
Only top level json keys are supported when providing input with query parameters.
Input query parameters are only read if the request has no body.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| workflow | `path` | string | `string` |  | ✓ |  | path to target workflow |
| ctype | `query` | string | `string` |  |  |  | Manually set the Content-Type response header instead of auto-detected. This doesn't change the body of the response in any way. |
| field | `query` | string | `string` |  |  |  | If provided, instead of returning the entire output json the response body will contain the single top-level json field |
| raw-output | `query` | boolean | `bool` |  |  |  | If set to true, will return an empty output as null, encoded base64 data as decoded binary data, and quoted json strings as a escaped string. |
| Workflow Input | `body` | [interface{}](#interface) | `interface{}` | |  | | The input of this workflow instance |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#await-execute-workflow-body-200) | OK | successfully executed workflow |  | [schema](#await-execute-workflow-body-200-schema) |

#### Responses


##### <span id="await-execute-workflow-body-200"></span> 200 - successfully executed workflow
Status: OK

###### <span id="await-execute-workflow-body-200-schema"></span> Schema

### <span id="broadcast-cloudevent"></span> Broadcast Cloud Event (*broadcastCloudevent*)

```
POST /api/namespaces/{namespace}/broadcast
```

Broadcast a cloud event to a namespace.
Cloud events posted to this api will be picked up by any workflows listening to the same event type on the namescape.
The body of this request should follow the cloud event core specification defined at https://github.com/cloudevents/spec .


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| cloudevent | `body` | [interface{}](#interface) | `interface{}` | |  | | Cloud Event request to be sent. |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#broadcast-cloudevent-200) | OK | successfully sent cloud event |  | [schema](#broadcast-cloudevent-200-schema) |

#### Responses


##### <span id="broadcast-cloudevent-200"></span> 200 - successfully sent cloud event
Status: OK

###### <span id="broadcast-cloudevent-200-schema"></span> Schema

### <span id="cancel-instance"></span> Cancel a Pending Instance (*cancelInstance*)

```
POST /api/namespaces/{namespace}/instances/{instance}/cancel
```

Cancel a currently pending instance.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| instance | `path` | string | `string` |  | ✓ |  | target instance |
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#cancel-instance-200) | OK | successfully cancelled instance |  | [schema](#cancel-instance-200-schema) |

#### Responses


##### <span id="cancel-instance-200"></span> 200 - successfully cancelled instance
Status: OK

###### <span id="cancel-instance-200-schema"></span> Schema

### <span id="create-directory"></span> Create a Directory (*createDirectory*)

```
PUT /api/namespaces/{namespace}/tree/{directory}?op=create-directory
```

Creates a directory at the target path.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| directory | `path` | string | `string` |  | ✓ |  | path to target directory |
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#create-directory-200) | OK | successfully created directory |  | [schema](#create-directory-200-schema) |

#### Responses


##### <span id="create-directory-200"></span> 200 - successfully created directory
Status: OK

###### <span id="create-directory-200-schema"></span> Schema

### <span id="create-global-service"></span> Create Global Service (*createGlobalService*)

```
POST /api/functions
```

Creates global scoped knative service.
Service Names are unique on a scope level.
These services can be used as functions in workflows, more about this can be read here:
https://docs.direktiv.io/docs/walkthrough/using-functions.html


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Service | `body` | [CreateGlobalServiceBody](#create-global-service-body) | `CreateGlobalServiceBody` | | ✓ | | Payload that contains information on new service |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#create-global-service-200) | OK | successfully created service |  | [schema](#create-global-service-200-schema) |

#### Responses


##### <span id="create-global-service-200"></span> 200 - successfully created service
Status: OK

###### <span id="create-global-service-200-schema"></span> Schema

###### Inlined models

**<span id="create-global-service-body"></span> CreateGlobalServiceBody**


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| cmd | string| `string` | ✓ | |  |  |
| image | string| `string` | ✓ | | Target image a service will use |  |
| minScale | integer| `int64` | ✓ | | Minimum amount of service pods to be live |  |
| name | string| `string` | ✓ | | Name of new service |  |
| size | string| `string` | ✓ | | Size of created service pods |  |



### <span id="create-namespace"></span> Creates a namespace (*createNamespace*)

```
PUT /api/namespaces/{namespace}
```

Creates a new namespace.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace to create |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#create-namespace-200) | OK | namespace has been successfully created |  | [schema](#create-namespace-200-schema) |

#### Responses


##### <span id="create-namespace-200"></span> 200 - namespace has been successfully created
Status: OK

###### <span id="create-namespace-200-schema"></span> Schema

### <span id="create-namespace-service"></span> Create Namespace Service (*createNamespaceService*)

```
POST /api/functions/namespaces/{namespace}
```

Creates namespace scoped knative service.
Service Names are unique on a scope level.
These services can be used as functions in workflows, more about this can be read here:
https://docs.direktiv.io/docs/walkthrough/using-functions.html


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| Service | `body` | [CreateNamespaceServiceBody](#create-namespace-service-body) | `CreateNamespaceServiceBody` | | ✓ | | Payload that contains information on new service |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#create-namespace-service-200) | OK | successfully created service |  | [schema](#create-namespace-service-200-schema) |

#### Responses


##### <span id="create-namespace-service-200"></span> 200 - successfully created service
Status: OK

###### <span id="create-namespace-service-200-schema"></span> Schema

###### Inlined models

**<span id="create-namespace-service-body"></span> CreateNamespaceServiceBody**


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| cmd | string| `string` | ✓ | |  |  |
| image | string| `string` | ✓ | | Target image a service will use |  |
| minScale | integer| `int64` | ✓ | | Minimum amount of service pods to be live |  |
| name | string| `string` | ✓ | | Name of new service |  |
| size | string| `string` | ✓ | | Size of created service pods |  |



### <span id="create-secret"></span> Create a Namespace Secret (*createSecret*)

```
PUT /api/namespaces/{namespace}/secrets/{secret}
```

Create a namespace secret.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| secret | `path` | string | `string` |  | ✓ |  | target secret |
| Secret Payload | `body` | string | `string` | |  | | Payload that contains secret data. |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#create-secret-200) | OK | successfully created namespace secret |  | [schema](#create-secret-200-schema) |

#### Responses


##### <span id="create-secret-200"></span> 200 - successfully created namespace secret
Status: OK

###### <span id="create-secret-200-schema"></span> Schema

### <span id="create-workflow"></span> Create a Workflow (*createWorkflow*)

```
PUT /api/namespaces/{namespace}/tree/{workflow}?op=create-workflow
```

Creates a workflow at the target path.
The body of this request should contain the workflow yaml.


#### Consumes
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| workflow | `path` | string | `string` |  | ✓ |  | path to target workflow |
| workflow data | `body` | string | `string` | |  | | Payload that contains the direktiv workflow yaml to create. |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#create-workflow-200) | OK | successfully created workflow |  | [schema](#create-workflow-200-schema) |

#### Responses


##### <span id="create-workflow-200"></span> 200 - successfully created workflow
Status: OK

###### <span id="create-workflow-200-schema"></span> Schema

### <span id="delete-global-revision"></span> Delete Global Service Revision (*deleteGlobalRevision*)

```
DELETE /api/functions/{serviceName}/revisions/{revisionGeneration}
```

Delete a global scoped knative service revision.
The target revision generation is the number suffix on a revision.
Example: A revisions named 'global-fast-request-00003' would have the revisionGeneration '00003'.
Note: Revisions with traffic cannot be deleted.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| revisionGeneration | `path` | string | `string` |  | ✓ |  | target revision generation |
| serviceName | `path` | string | `string` |  | ✓ |  | target service name |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#delete-global-revision-200) | OK | successfully deleted service revision |  | [schema](#delete-global-revision-200-schema) |

#### Responses


##### <span id="delete-global-revision-200"></span> 200 - successfully deleted service revision
Status: OK

###### <span id="delete-global-revision-200-schema"></span> Schema

### <span id="delete-global-service"></span> Delete Global Service (*deleteGlobalService*)

```
DELETE /api/functions/{serviceName}
```

Deletes global scoped knative service and all its revisions.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| serviceName | `path` | string | `string` |  | ✓ |  | target service name |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#delete-global-service-200) | OK | successfully deleted service |  | [schema](#delete-global-service-200-schema) |

#### Responses


##### <span id="delete-global-service-200"></span> 200 - successfully deleted service
Status: OK

###### <span id="delete-global-service-200-schema"></span> Schema

### <span id="delete-instance-variable"></span> Delete a Instance Variable (*deleteInstanceVariable*)

```
DELETE /api/namespaces/{namespace}/instances/{instance}/vars/{variable}
```

Delete a instance variable.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| instance | `path` | string | `string` |  | ✓ |  | target instance |
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| variable | `path` | string | `string` |  | ✓ |  | target variable |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#delete-instance-variable-200) | OK | successfully deleted instance variable |  | [schema](#delete-instance-variable-200-schema) |

#### Responses


##### <span id="delete-instance-variable-200"></span> 200 - successfully deleted instance variable
Status: OK

###### <span id="delete-instance-variable-200-schema"></span> Schema

### <span id="delete-namespace"></span> Delete a namespace (*deleteNamespace*)

```
DELETE /api/namespaces/{namespace}
```

Delete a namespace.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace to delete |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#delete-namespace-200) | OK | namespace has been successfully deleted |  | [schema](#delete-namespace-200-schema) |

#### Responses


##### <span id="delete-namespace-200"></span> 200 - namespace has been successfully deleted
Status: OK

###### <span id="delete-namespace-200-schema"></span> Schema

### <span id="delete-namespace-revision"></span> Delete Namespace Service Revision (*deleteNamespaceRevision*)

```
DELETE /api/functions/namespaces/{namespace}/function/{serviceName}/revisions/{revisionGeneration}
```

Delete a namespace scoped knative service revision.
The target revision generation is the number suffix on a revision.
Example: A revisions named 'namespace-direktiv-fast-request-00003' would have the revisionGeneration '00003'.
Note: Revisions with traffic cannot be deleted.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| revisionGeneration | `path` | string | `string` |  | ✓ |  | target revision generation |
| serviceName | `path` | string | `string` |  | ✓ |  | target service name |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#delete-namespace-revision-200) | OK | successfully deleted service revision |  | [schema](#delete-namespace-revision-200-schema) |

#### Responses


##### <span id="delete-namespace-revision-200"></span> 200 - successfully deleted service revision
Status: OK

###### <span id="delete-namespace-revision-200-schema"></span> Schema

### <span id="delete-namespace-service"></span> Delete Namespace Service (*deleteNamespaceService*)

```
DELETE /api/functions/namespaces/{namespace}/function/{serviceName}
```

Deletes namespace scoped knative service and all its revisions.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| serviceName | `path` | string | `string` |  | ✓ |  | target service name |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#delete-namespace-service-200) | OK | successfully deleted service |  | [schema](#delete-namespace-service-200-schema) |

#### Responses


##### <span id="delete-namespace-service-200"></span> 200 - successfully deleted service
Status: OK

###### <span id="delete-namespace-service-200-schema"></span> Schema

### <span id="delete-namespace-variable"></span> Delete a Namespace Variable (*deleteNamespaceVariable*)

```
DELETE /api/namespaces/{namespace}/vars/{variable}
```

Delete a namespace variable.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| variable | `path` | string | `string` |  | ✓ |  | target variable |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#delete-namespace-variable-200) | OK | successfully deleted namespace variable |  | [schema](#delete-namespace-variable-200-schema) |

#### Responses


##### <span id="delete-namespace-variable-200"></span> 200 - successfully deleted namespace variable
Status: OK

###### <span id="delete-namespace-variable-200-schema"></span> Schema

### <span id="delete-registry"></span> Delete a Namespace Container Registry (*deleteRegistry*)

```
POST /api/namespaces/{namespace}/registries
```

Delete a namespace container registry


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| Registry Payload | `body` | [DeleteRegistryBody](#delete-registry-body) | `DeleteRegistryBody` | |  | | Payload that contains registry data |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#delete-registry-200) | OK | successfully delete namespace registry |  | [schema](#delete-registry-200-schema) |

#### Responses


##### <span id="delete-registry-200"></span> 200 - successfully delete namespace registry
Status: OK

###### <span id="delete-registry-200-schema"></span> Schema

###### Inlined models

**<span id="delete-registry-body"></span> DeleteRegistryBody**


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| reg | string| `string` | ✓ | | Target registry URL |  |



### <span id="delete-secret"></span> Delete a Namespace Secret (*deleteSecret*)

```
DELETE /api/namespaces/{namespace}/secrets/{secret}
```

Delete a namespace secret.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| secret | `path` | string | `string` |  | ✓ |  | target secret |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#delete-secret-200) | OK | successfully deleted namespace secret |  | [schema](#delete-secret-200-schema) |

#### Responses


##### <span id="delete-secret-200"></span> 200 - successfully deleted namespace secret
Status: OK

###### <span id="delete-secret-200-schema"></span> Schema

### <span id="delete-workflow-variable"></span> Delete a Workflow Variable (*deleteWorkflowVariable*)

```
DELETE /api/namespaces/{namespace}/tree/{workflow}?op=delete-var
```

Delete a workflow variable.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| workflow | `path` | string | `string` |  | ✓ |  | path to target workflow |
| var | `query` | string | `string` |  | ✓ |  | target variable |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#delete-workflow-variable-200) | OK | successfully deleted workflow variable |  | [schema](#delete-workflow-variable-200-schema) |

#### Responses


##### <span id="delete-workflow-variable-200"></span> 200 - successfully deleted workflow variable
Status: OK

###### <span id="delete-workflow-variable-200-schema"></span> Schema

### <span id="execute-workflow"></span> Execute a Workflow (*executeWorkflow*)

```
POST /api/namespaces/{namespace}/tree/{workflow}?op=execute
```

Executes a workflow with optionally some input provided in the request body as json.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| workflow | `path` | string | `string` |  | ✓ |  | path to target workflow |
| Workflow Input | `body` | [interface{}](#interface) | `interface{}` | |  | | The input of this workflow instance |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#execute-workflow-200) | OK | successfully executed workflow |  | [schema](#execute-workflow-200-schema) |

#### Responses


##### <span id="execute-workflow-200"></span> 200 - successfully executed workflow
Status: OK

###### <span id="execute-workflow-200-schema"></span> Schema

### <span id="get-global-service"></span> Get Global Service Details (*getGlobalService*)

```
GET /api/functions/{serviceName}
```

Get details of a global scoped knative service.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| serviceName | `path` | string | `string` |  | ✓ |  | target service name |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-global-service-200) | OK | successfully got service details |  | [schema](#get-global-service-200-schema) |

#### Responses


##### <span id="get-global-service-200"></span> 200 - successfully got service details
Status: OK

###### <span id="get-global-service-200-schema"></span> Schema

### <span id="get-global-service-list"></span> Get Global Services List (*getGlobalServiceList*)

```
GET /api/functions
```

Gets a list of global knative services.


#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-global-service-list-200) | OK | successfully got services list |  | [schema](#get-global-service-list-200-schema) |

#### Responses


##### <span id="get-global-service-list-200"></span> 200 - successfully got services list
Status: OK

###### <span id="get-global-service-list-200-schema"></span> Schema

### <span id="get-instance"></span> Get a Instance (*getInstance*)

```
GET /api/namespaces/{namespace}/instances/{instance}
```

Gets the details of a executed workflow instance in this namespace.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| instance | `path` | string | `string` |  | ✓ |  | target instance |
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-instance-200) | OK | successfully got instance |  | [schema](#get-instance-200-schema) |

#### Responses


##### <span id="get-instance-200"></span> 200 - successfully got instance
Status: OK

###### <span id="get-instance-200-schema"></span> Schema

### <span id="get-instance-input"></span> Get a Instance Input (*getInstanceInput*)

```
GET /api/namespaces/{namespace}/instances/{instance}/input
```

Gets the input an instance was provided when executed.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| instance | `path` | string | `string` |  | ✓ |  | target instance |
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-instance-input-200) | OK | successfully got instance input |  | [schema](#get-instance-input-200-schema) |

#### Responses


##### <span id="get-instance-input-200"></span> 200 - successfully got instance input
Status: OK

###### <span id="get-instance-input-200-schema"></span> Schema

### <span id="get-instance-list"></span> Get List Instances (*getInstanceList*)

```
GET /api/namespaces/{namespace}/instances
```

Gets a list of instances in a namespace.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-instance-list-200) | OK | successfully got namespace instances |  | [schema](#get-instance-list-200-schema) |

#### Responses


##### <span id="get-instance-list-200"></span> 200 - successfully got namespace instances
Status: OK

###### <span id="get-instance-list-200-schema"></span> Schema

### <span id="get-instance-output"></span> Get a Instance Output (*getInstanceOutput*)

```
GET /api/namespaces/{namespace}/instances/{instance}/output
```

Gets the output an instance was provided when executed.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| instance | `path` | string | `string` |  | ✓ |  | target instance |
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-instance-output-200) | OK | successfully got instance output |  | [schema](#get-instance-output-200-schema) |

#### Responses


##### <span id="get-instance-output-200"></span> 200 - successfully got instance output
Status: OK

###### <span id="get-instance-output-200-schema"></span> Schema

### <span id="get-instance-variable"></span> Get a Instance Variable (*getInstanceVariable*)

```
GET /api/namespaces/{namespace}/instances/{instance}/vars/{variable}
```

Get the value sorted in a instance variable.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| instance | `path` | string | `string` |  | ✓ |  | target instance |
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| variable | `path` | string | `string` |  | ✓ |  | target variable |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-instance-variable-200) | OK | successfully got instance variable |  | [schema](#get-instance-variable-200-schema) |

#### Responses


##### <span id="get-instance-variable-200"></span> 200 - successfully got instance variable
Status: OK

###### <span id="get-instance-variable-200-schema"></span> Schema

### <span id="get-instance-variables"></span> Get List of Instance Variable (*getInstanceVariables*)

```
GET /api/namespaces/{namespace}/instances/{instance}/vars
```

Gets a list of variables in a instance.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| instance | `path` | int32 (formatted string) | `string` |  | ✓ |  | target instance |
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-instance-variables-200) | OK | successfully got instance variables |  | [schema](#get-instance-variables-200-schema) |

#### Responses


##### <span id="get-instance-variables-200"></span> 200 - successfully got instance variables
Status: OK

###### <span id="get-instance-variables-200-schema"></span> Schema

### <span id="get-namespace-service"></span> Get Namespace Service Details (*getNamespaceService*)

```
GET /api/functions/namespaces/{namespace}/function/{serviceName}
```

Get details of a namespace scoped knative service.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| serviceName | `path` | string | `string` |  | ✓ |  | target service name |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-namespace-service-200) | OK | successfully got service details |  | [schema](#get-namespace-service-200-schema) |

#### Responses


##### <span id="get-namespace-service-200"></span> 200 - successfully got service details
Status: OK

###### <span id="get-namespace-service-200-schema"></span> Schema

### <span id="get-namespace-service-list"></span> Get Namespace Services List (*getNamespaceServiceList*)

```
GET /api/functions/namespaces/{namespace}
```

Gets a list of namespace knative services.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-namespace-service-list-200) | OK | successfully got services list |  | [schema](#get-namespace-service-list-200-schema) |

#### Responses


##### <span id="get-namespace-service-list-200"></span> 200 - successfully got services list
Status: OK

###### <span id="get-namespace-service-list-200-schema"></span> Schema

### <span id="get-namespace-variable"></span> Get a Namespace Variable (*getNamespaceVariable*)

```
GET /api/namespaces/{namespace}/vars/{variable}
```

Get the value sorted in a namespace variable.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| variable | `path` | string | `string` |  | ✓ |  | target variable |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-namespace-variable-200) | OK | successfully got namespace variable |  | [schema](#get-namespace-variable-200-schema) |

#### Responses


##### <span id="get-namespace-variable-200"></span> 200 - successfully got namespace variable
Status: OK

###### <span id="get-namespace-variable-200-schema"></span> Schema

### <span id="get-namespace-variables"></span> Get Namespace Variable List (*getNamespaceVariables*)

```
GET /api/namespaces/{namespace}/vars
```

Gets a list of variables in a namespace.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-namespace-variables-200) | OK | successfully got namespace variables |  | [schema](#get-namespace-variables-200-schema) |

#### Responses


##### <span id="get-namespace-variables-200"></span> 200 - successfully got namespace variables
Status: OK

###### <span id="get-namespace-variables-200-schema"></span> Schema

### <span id="get-namespaces"></span> Gets the list of namespaces (*getNamespaces*)

```
GET /api/namespaces
```

Gets the list of namespaces.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| after | `query` | string | `string` |  |  |  |  |
| before | `query` | string | `string` |  |  |  |  |
| filter.field | `query` | string | `string` |  |  |  |  |
| filter.type | `query` | string | `string` |  |  |  |  |
| filter.val | `query` | string | `string` |  |  |  |  |
| first | `query` | int32 (formatted integer) | `int32` |  |  |  |  |
| last | `query` | int32 (formatted integer) | `int32` |  |  |  |  |
| order.direction | `query` | string | `string` |  |  |  |  |
| order.field | `query` | string | `string` |  |  |  |  |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-namespaces-200) | OK | successfully got list of namespaces |  | [schema](#get-namespaces-200-schema) |

#### Responses


##### <span id="get-namespaces-200"></span> 200 - successfully got list of namespaces
Status: OK

###### <span id="get-namespaces-200-schema"></span> Schema

### <span id="get-nodes"></span> Get List of Namespace Nodes (*getNodes*)

```
GET /api/namespaces/{namespace}/tree/{nodePath}
```

Gets Workflow and Directory Nodes at nodePath.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| nodePath | `path` | int32 (formatted string) | `string` |  | ✓ |  | target path in tree |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-nodes-200) | OK | successfully got namespace nodes |  | [schema](#get-nodes-200-schema) |

#### Responses


##### <span id="get-nodes-200"></span> 200 - successfully got namespace nodes
Status: OK

###### <span id="get-nodes-200-schema"></span> Schema

### <span id="get-registries"></span> Get List of Namespace Registries (*getRegistries*)

```
GET /api/namespaces/{namespace}/registries
```

Gets the list of namespace registries.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-registries-200) | OK | successfully got namespace registries |  | [schema](#get-registries-200-schema) |

#### Responses


##### <span id="get-registries-200"></span> 200 - successfully got namespace registries
Status: OK

###### <span id="get-registries-200-schema"></span> Schema

### <span id="get-secrets"></span> Get List of Namespace Secrets (*getSecrets*)

```
GET /api/namespaces/{namespace}/secrets
```

Gets the list of namespace secrets.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-secrets-200) | OK | successfully got namespace secrets |  | [schema](#get-secrets-200-schema) |

#### Responses


##### <span id="get-secrets-200"></span> 200 - successfully got namespace secrets
Status: OK

###### <span id="get-secrets-200-schema"></span> Schema

### <span id="get-workflow-logs"></span> Get Workflow Level Logs (*getWorkflowLogs*)

```
GET /api/namespaces/{namespace}/tree/{workflow}?op=logs
```

Get workflow level logs.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| workflow | `path` | string | `string` |  | ✓ |  | path to target workflow |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-workflow-logs-200) | OK | successfully got workflow logs |  | [schema](#get-workflow-logs-200-schema) |

#### Responses


##### <span id="get-workflow-logs-200"></span> 200 - successfully got workflow logs
Status: OK

###### <span id="get-workflow-logs-200-schema"></span> Schema

### <span id="get-workflow-service"></span> Get Workflow Service Details (*getWorkflowService*)

```
GET /api/functions/namespaces/{namespace}/tree/{workflow}?op=function
```

Get a workflow scoped knative service details.
Note: This is a Server-Sent-Event endpoint, and will not work with the default swagger client.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| workflow | `path` | string | `string` |  | ✓ |  | path to target workflow |
| svn | `query` | string | `string` |  | ✓ |  | target service name |
| version | `query` | string | `string` |  | ✓ |  | target service version |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-workflow-service-200) | OK | successfully got service details |  | [schema](#get-workflow-service-200-schema) |

#### Responses


##### <span id="get-workflow-service-200"></span> 200 - successfully got service details
Status: OK

###### <span id="get-workflow-service-200-schema"></span> Schema

### <span id="get-workflow-service-revision"></span> Get Workflow Service Revision (*getWorkflowServiceRevision*)

```
GET /api/functions/namespaces/{namespace}/tree/{workflow}?op=function-revision
```

Get a workflow scoped knative service revision.
This will return details on a single revision.
The target revision generation (rev query) is the number suffix on a revision.
Example: A revisions named 'workflow-10640097968065193909-get-00001' would have the revisionGeneration '00001'.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| workflow | `path` | string | `string` |  | ✓ |  | path to target workflow |
| rev | `query` | string | `string` |  | ✓ |  | target service revison |
| svn | `query` | string | `string` |  | ✓ |  | target service name |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-workflow-service-revision-200) | OK | successfully got service revision details |  | [schema](#get-workflow-service-revision-200-schema) |

#### Responses


##### <span id="get-workflow-service-revision-200"></span> 200 - successfully got service revision details
Status: OK

###### <span id="get-workflow-service-revision-200-schema"></span> Schema

### <span id="get-workflow-service-revision-list"></span> Get Workflow Service Revision List (*getWorkflowServiceRevisionList*)

```
GET /api/functions/namespaces/{namespace}/tree/{workflow}?op=function-revisions
```

Get the revision list of a workflow scoped knative service.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| workflow | `path` | string | `string` |  | ✓ |  | path to target workflow |
| svn | `query` | string | `string` |  | ✓ |  | target service name |
| version | `query` | string | `string` |  | ✓ |  | target service version |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-workflow-service-revision-list-200) | OK | successfully got service revisions |  | [schema](#get-workflow-service-revision-list-200-schema) |

#### Responses


##### <span id="get-workflow-service-revision-list-200"></span> 200 - successfully got service revisions
Status: OK

###### <span id="get-workflow-service-revision-list-200-schema"></span> Schema

### <span id="get-workflow-variable"></span> Get a Workflow Variable (*getWorkflowVariable*)

```
GET /api/namespaces/{namespace}/tree/{workflow}?op=var
```

Get the value sorted in a workflow variable.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| workflow | `path` | string | `string` |  | ✓ |  | path to target workflow |
| var | `query` | string | `string` |  | ✓ |  | target variable |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-workflow-variable-200) | OK | successfully got workflow variable |  | [schema](#get-workflow-variable-200-schema) |

#### Responses


##### <span id="get-workflow-variable-200"></span> 200 - successfully got workflow variable
Status: OK

###### <span id="get-workflow-variable-200-schema"></span> Schema

### <span id="get-workflow-variables"></span> Get List of Workflow Variables (*getWorkflowVariables*)

```
GET /api/namespaces/{namespace}/tree/{workflow}?op=vars
```

Gets a list of variables in a workflow.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| workflow | `path` | int32 (formatted string) | `string` |  | ✓ |  | path to target workflow |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-workflow-variables-200) | OK | successfully got workflow variables |  | [schema](#get-workflow-variables-200-schema) |

#### Responses


##### <span id="get-workflow-variables-200"></span> 200 - successfully got workflow variables
Status: OK

###### <span id="get-workflow-variables-200-schema"></span> Schema

### <span id="instance-logs"></span> Gets Instance Logs (*instanceLogs*)

```
GET /api/namespaces/{namespace}/instances/{instance}/logs
```

Gets the logs of an executed instance.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| instance | `path` | int32 (formatted string) | `string` |  | ✓ |  | target instance id |
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#instance-logs-200) | OK | successfully got instance logs |  | [schema](#instance-logs-200-schema) |

#### Responses


##### <span id="instance-logs-200"></span> 200 - successfully got instance logs
Status: OK

###### <span id="instance-logs-200-schema"></span> Schema

### <span id="jq-playground"></span> JQ Playground api to test jq queries (*jqPlayground*)

```
POST /api/jq
```

JQ Playground is a sandbox where you can test jq queries with custom data.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| JQ payload | `body` | [JqPlaygroundBody](#jq-playground-body) | `JqPlaygroundBody` | |  | | Payload that contains both the JSON data to manipulate and jq query. |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#jq-playground-200) | OK | jq query was successful |  | [schema](#jq-playground-200-schema) |

#### Responses


##### <span id="jq-playground-200"></span> 200 - jq query was successful
Status: OK

###### <span id="jq-playground-200-schema"></span> Schema

###### Inlined models

**<span id="jq-playground-body"></span> JqPlaygroundBody**


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | string| `string` | ✓ | | JSON data encoded in base64 |  |
| query | string| `string` | ✓ | | jq query to manipulate JSON data |  |



### <span id="list-global-service-revision-pods"></span> Get Global Service Revision Pods List (*listGlobalServiceRevisionPods*)

```
GET /api/functions/{serviceName}/revisions/{revisionGeneration}/pods
```

List a revisions pods of a global scoped knative service.
The target revision generation is the number suffix on a revision.
Example: A revisions named 'global-fast-request-00003' would have the revisionGeneration '00003' .


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| revisionGeneration | `path` | string | `string` |  | ✓ |  | target revision generation |
| serviceName | `path` | string | `string` |  | ✓ |  | target service name |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#list-global-service-revision-pods-200) | OK | successfully got list of a service revision pods |  | [schema](#list-global-service-revision-pods-200-schema) |

#### Responses


##### <span id="list-global-service-revision-pods-200"></span> 200 - successfully got list of a service revision pods
Status: OK

###### <span id="list-global-service-revision-pods-200-schema"></span> Schema

### <span id="list-namespace-service-revision-pods"></span> Get Namespace Service Revision Pods List (*listNamespaceServiceRevisionPods*)

```
GET /api/functions/namespaces/{namespace}/function/{serviceName}/revisions/{revisionGeneration}/pods
```

List a revisions pods of a namespace scoped knative service.
The target revision generation is the number suffix on a revision.
Example: A revisions named 'namespace-direktiv-fast-request-00003' would have the revisionGeneration '00003'.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| revisionGeneration | `path` | string | `string` |  | ✓ |  | target revision generation |
| serviceName | `path` | string | `string` |  | ✓ |  | target service name |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#list-namespace-service-revision-pods-200) | OK | successfully got list of a service revision pods |  | [schema](#list-namespace-service-revision-pods-200-schema) |

#### Responses


##### <span id="list-namespace-service-revision-pods-200"></span> 200 - successfully got list of a service revision pods
Status: OK

###### <span id="list-namespace-service-revision-pods-200-schema"></span> Schema

### <span id="list-workflow-service-revision-pods"></span> Get Workflow Service Revision Pods List (*listWorkflowServiceRevisionPods*)

```
GET /api/functions/namespaces/{namespace}/tree/{workflow}?op=pods
```

List a revisions pods of a workflow scoped knative service.
The target revision generation (rev query) is the number suffix on a revision.
Example: A revisions named 'workflow-10640097968065193909-get-00001' would have the revisionGeneration '00001'.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| workflow | `path` | string | `string` |  | ✓ |  | path to target workflow |
| rev | `query` | string | `string` |  | ✓ |  | target service revison |
| svn | `query` | string | `string` |  | ✓ |  | target service name |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#list-workflow-service-revision-pods-200) | OK | successfully got list of a service revision pods |  | [schema](#list-workflow-service-revision-pods-200-schema) |

#### Responses


##### <span id="list-workflow-service-revision-pods-200"></span> 200 - successfully got list of a service revision pods
Status: OK

###### <span id="list-workflow-service-revision-pods-200-schema"></span> Schema

### <span id="list-workflow-services"></span> Get Workflow Services List (*listWorkflowServices*)

```
GET /api/functions/namespaces/{namespace}/tree/{workflow}?op=services
```

Gets a list of workflow knative services.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| workflow | `path` | string | `string` |  | ✓ |  | path to target workflow |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#list-workflow-services-200) | OK | successfully got services list |  | [schema](#list-workflow-services-200-schema) |

#### Responses


##### <span id="list-workflow-services-200"></span> 200 - successfully got services list
Status: OK

###### <span id="list-workflow-services-200-schema"></span> Schema

### <span id="namespace-logs"></span> Gets Namespace Level Logs (*namespaceLogs*)

```
GET /api/namespaces/{namespace}/logs
```

Gets Namespace Level Logs.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#namespace-logs-200) | OK | successfully got namespace logs |  | [schema](#namespace-logs-200-schema) |

#### Responses


##### <span id="namespace-logs-200"></span> 200 - successfully got namespace logs
Status: OK

###### <span id="namespace-logs-200-schema"></span> Schema

### <span id="namespace-metrics-failed"></span> Gets Namespace Failed Workflow Instances Metrics (*namespaceMetricsFailed*)

```
GET /api/namespaces/{namespace}/metrics/failed
```

Get metrics of failed workflows in the targeted namespace.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#namespace-metrics-failed-200) | OK | successfully got namespace metrics |  | [schema](#namespace-metrics-failed-200-schema) |

#### Responses


##### <span id="namespace-metrics-failed-200"></span> 200 - successfully got namespace metrics
Status: OK

###### <span id="namespace-metrics-failed-200-schema"></span> Schema

### <span id="namespace-metrics-invoked"></span> Gets Namespace Invoked Workflow Metrics (*namespaceMetricsInvoked*)

```
GET /api/namespaces/{namespace}/metrics/invoked
```

Get metrics of invoked workflows in the targeted namespace.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#namespace-metrics-invoked-200) | OK | successfully got namespace metrics |  | [schema](#namespace-metrics-invoked-200-schema) |

#### Responses


##### <span id="namespace-metrics-invoked-200"></span> 200 - successfully got namespace metrics
Status: OK

###### <span id="namespace-metrics-invoked-200-schema"></span> Schema

### <span id="namespace-metrics-milliseconds"></span> Gets Namespace Workflow Timing Metrics (*namespaceMetricsMilliseconds*)

```
GET /api/namespaces/{namespace}/metrics/milliseconds
```

Get timing metrics of workflows in the targeted namespace.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#namespace-metrics-milliseconds-200) | OK | successfully got namespace metrics |  | [schema](#namespace-metrics-milliseconds-200-schema) |

#### Responses


##### <span id="namespace-metrics-milliseconds-200"></span> 200 - successfully got namespace metrics
Status: OK

###### <span id="namespace-metrics-milliseconds-200-schema"></span> Schema

### <span id="namespace-metrics-successful"></span> Gets Namespace Successful Workflow Instances Metrics (*namespaceMetricsSuccessful*)

```
GET /api/namespaces/{namespace}/metrics/successful
```

Get metrics of successful workflows in the targeted namespace.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#namespace-metrics-successful-200) | OK | successfully got namespace metrics |  | [schema](#namespace-metrics-successful-200-schema) |

#### Responses


##### <span id="namespace-metrics-successful-200"></span> 200 - successfully got namespace metrics
Status: OK

###### <span id="namespace-metrics-successful-200-schema"></span> Schema

### <span id="server-logs"></span> Get Direktiv Server Logs (*serverLogs*)

```
GET /api/logs
```

Gets Direktiv Server Logs.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| after | `query` | string | `string` |  |  |  |  |
| before | `query` | string | `string` |  |  |  |  |
| filter.field | `query` | string | `string` |  |  |  |  |
| filter.type | `query` | string | `string` |  |  |  |  |
| filter.val | `query` | string | `string` |  |  |  |  |
| first | `query` | int32 (formatted integer) | `int32` |  |  |  |  |
| last | `query` | int32 (formatted integer) | `int32` |  |  |  |  |
| order.direction | `query` | string | `string` |  |  |  |  |
| order.field | `query` | string | `string` |  |  |  |  |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#server-logs-200) | OK | successfully got server logs |  | [schema](#server-logs-200-schema) |

#### Responses


##### <span id="server-logs-200"></span> 200 - successfully got server logs
Status: OK

###### <span id="server-logs-200-schema"></span> Schema

### <span id="set-instance-variable"></span> Set a Instance Variable (*setInstanceVariable*)

```
PUT /api/namespaces/{namespace}/instances/{instance}/vars/{variable}
```

Set the value sorted in a instance variable.
If the target variable does not exists, it will be created.
Variable data can be anything so long as it can be represented as a string.


#### Consumes
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| instance | `path` | string | `string` |  | ✓ |  | target instance |
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| variable | `path` | string | `string` |  | ✓ |  | target variable |
| data | `body` | string | `string` | |  | | Payload that contains variable data. |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#set-instance-variable-200) | OK | successfully set instance variable |  | [schema](#set-instance-variable-200-schema) |

#### Responses


##### <span id="set-instance-variable-200"></span> 200 - successfully set instance variable
Status: OK

###### <span id="set-instance-variable-200-schema"></span> Schema

### <span id="set-namespace-variable"></span> Set a Namespace Variable (*setNamespaceVariable*)

```
PUT /api/namespaces/{namespace}/vars/{variable}
```

Set the value sorted in a namespace variable.
If the target variable does not exists, it will be created.
Variable data can be anything so long as it can be represented as a string.


#### Consumes
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| variable | `path` | string | `string` |  | ✓ |  | target variable |
| data | `body` | string | `string` | |  | | Payload that contains variable data. |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#set-namespace-variable-200) | OK | successfully set namespace variable |  | [schema](#set-namespace-variable-200-schema) |

#### Responses


##### <span id="set-namespace-variable-200"></span> 200 - successfully set namespace variable
Status: OK

###### <span id="set-namespace-variable-200-schema"></span> Schema

### <span id="set-workflow-cloud-event-logs"></span> Set Cloud Event for Workflow to Log to (*setWorkflowCloudEventLogs*)

```
POST /api/namespaces/{namespace}/tree/{workflow}?op=set-workflow-event-logging
```

Set Cloud Event for Workflow to Log to.
When configured type `direktiv.instanceLog` cloud events will be generated with the `logger` parameter set to the configured value.
Workflows can be configured to generate cloud events on their namespace anything the log parameter produces data.
Please find more information on this topic here:
https://docs.direktiv.io/docs/examples/logging.html


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| workflow | `path` | string | `string` |  | ✓ |  | path to target workflow |
| Cloud Event Logger | `body` | [SetWorkflowCloudEventLogsBody](#set-workflow-cloud-event-logs-body) | `SetWorkflowCloudEventLogsBody` | |  | | Cloud event logger to target |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#set-workflow-cloud-event-logs-200) | OK | successfully update workflow |  | [schema](#set-workflow-cloud-event-logs-200-schema) |

#### Responses


##### <span id="set-workflow-cloud-event-logs-200"></span> 200 - successfully update workflow
Status: OK

###### <span id="set-workflow-cloud-event-logs-200-schema"></span> Schema

###### Inlined models

**<span id="set-workflow-cloud-event-logs-body"></span> SetWorkflowCloudEventLogsBody**


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| logger | string| `string` | ✓ | | Target Cloud Event |  |



### <span id="set-workflow-variable"></span> Set a Workflow Variable (*setWorkflowVariable*)

```
PUT /api/namespaces/{namespace}/tree/{workflow}?op=set-var
```

Set the value sorted in a workflow variable.
If the target variable does not exists, it will be created.
Variable data can be anything so long as it can be represented as a string.


#### Consumes
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| workflow | `path` | string | `string` |  | ✓ |  | path to target workflow |
| var | `query` | string | `string` |  | ✓ |  | target variable |
| data | `body` | string | `string` | |  | | Payload that contains variable data. |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#set-workflow-variable-200) | OK | successfully set workflow variable |  | [schema](#set-workflow-variable-200-schema) |

#### Responses


##### <span id="set-workflow-variable-200"></span> 200 - successfully set workflow variable
Status: OK

###### <span id="set-workflow-variable-200-schema"></span> Schema

### <span id="toggle-workflow"></span> Set Cloud Event for Workflow to Log to (*toggleWorkflow*)

```
POST /api/namespaces/{namespace}/tree/{workflow}?op=toggle
```

Toggle's whether or not a workflow is active.
Disabled workflows cannot be invoked. This includes start event and scheduled workflows.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| workflow | `path` | string | `string` |  | ✓ |  | path to target workflow |
| Workflow Live Status | `body` | [ToggleWorkflowBody](#toggle-workflow-body) | `ToggleWorkflowBody` | |  | | Whether or not the workflow is alive or disabled |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#toggle-workflow-200) | OK | successfully updated workflow live status |  | [schema](#toggle-workflow-200-schema) |

#### Responses


##### <span id="toggle-workflow-200"></span> 200 - successfully updated workflow live status
Status: OK

###### <span id="toggle-workflow-200-schema"></span> Schema

###### Inlined models

**<span id="toggle-workflow-body"></span> ToggleWorkflowBody**


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| live | boolean| `bool` | ✓ | | Workflow live status |  |



### <span id="update-global-service"></span> Create Global Service Revision (*updateGlobalService*)

```
POST /api/functions/{serviceName}
```

Creates a new global scoped knative service revision
Revisions are created with a traffic percentage. This percentage controls how much traffic will be directed to this revision.
Traffic can be set to 100 to direct all traffic.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| serviceName | `path` | string | `string` |  | ✓ |  | target service name |
| Service | `body` | [UpdateGlobalServiceBody](#update-global-service-body) | `UpdateGlobalServiceBody` | | ✓ | | Payload that contains information on service revision |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#update-global-service-200) | OK | successfully created service revision |  | [schema](#update-global-service-200-schema) |

#### Responses


##### <span id="update-global-service-200"></span> 200 - successfully created service revision
Status: OK

###### <span id="update-global-service-200-schema"></span> Schema

###### Inlined models

**<span id="update-global-service-body"></span> UpdateGlobalServiceBody**


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| cmd | string| `string` | ✓ | |  |  |
| image | string| `string` | ✓ | | Target image a service will use |  |
| minScale | integer| `int64` | ✓ | | Minimum amount of service pods to be live |  |
| size | string| `string` | ✓ | | Size of created service pods |  |
| trafficPercent | integer| `int64` | ✓ | | Traffic percentage new revision will use |  |



### <span id="update-global-service-traffic"></span> Update Global Service Traffic (*updateGlobalServiceTraffic*)

```
PATCH /api/functions/{serviceName}
```

Update Global Service traffic directed to each revision, traffic can only be configured between two revisions.
All other revisions will bet set to 0 traffic.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| serviceName | `path` | string | `string` |  | ✓ |  | target service name |
| Service Traffic | `body` | [UpdateGlobalServiceTrafficBody](#update-global-service-traffic-body) | `UpdateGlobalServiceTrafficBody` | | ✓ | | Payload that contains information on service traffic |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#update-global-service-traffic-200) | OK | successfully updated service traffic |  | [schema](#update-global-service-traffic-200-schema) |

#### Responses


##### <span id="update-global-service-traffic-200"></span> 200 - successfully updated service traffic
Status: OK

###### <span id="update-global-service-traffic-200-schema"></span> Schema

###### Inlined models

**<span id="update-global-service-traffic-body"></span> UpdateGlobalServiceTrafficBody**


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| values | [][UpdateGlobalServiceTrafficParamsBodyValuesItems0](#update-global-service-traffic-params-body-values-items0)| `[]*UpdateGlobalServiceTrafficParamsBodyValuesItems0` | ✓ | | List of revision traffic targets |  |



**<span id="update-global-service-traffic-params-body-values-items0"></span> UpdateGlobalServiceTrafficParamsBodyValuesItems0**


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| percent | integer| `int64` |  | | Target traffice percentage |  |
| revision | string| `string` |  | | Target service revision |  |



### <span id="update-namespace-service"></span> Create Namespace Service Revision (*updateNamespaceService*)

```
POST /api/functions/namespaces/{namespace}/function/{serviceName}
```

Creates a new namespace scoped knative service revision.
Revisions are created with a traffic percentage. This percentage controls
how much traffic will be directed to this revision. Traffic can be set to 100
to direct all traffic.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| serviceName | `path` | string | `string` |  | ✓ |  | target service name |
| Service | `body` | [UpdateNamespaceServiceBody](#update-namespace-service-body) | `UpdateNamespaceServiceBody` | | ✓ | | Payload that contains information on service revision |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#update-namespace-service-200) | OK | successfully created service revision |  | [schema](#update-namespace-service-200-schema) |

#### Responses


##### <span id="update-namespace-service-200"></span> 200 - successfully created service revision
Status: OK

###### <span id="update-namespace-service-200-schema"></span> Schema

###### Inlined models

**<span id="update-namespace-service-body"></span> UpdateNamespaceServiceBody**


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| cmd | string| `string` | ✓ | |  |  |
| image | string| `string` | ✓ | | Target image a service will use |  |
| minScale | integer| `int64` | ✓ | | Minimum amount of service pods to be live |  |
| size | string| `string` | ✓ | | Size of created service pods |  |
| trafficPercent | integer| `int64` | ✓ | | Traffic percentage new revision will use |  |



### <span id="update-namespace-service-traffic"></span> Update Namespace Service Traffic (*updateNamespaceServiceTraffic*)

```
PATCH /api/functions/namespaces/{namespace}/function/{serviceName}
```

Update Namespace Service traffic directed to each revision,
traffic can only be configured between two revisions. All other revisions
will bet set to 0 traffic.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| serviceName | `path` | string | `string` |  | ✓ |  | target service name |
| Service Traffic | `body` | [UpdateNamespaceServiceTrafficBody](#update-namespace-service-traffic-body) | `UpdateNamespaceServiceTrafficBody` | | ✓ | | Payload that contains information on service traffic |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#update-namespace-service-traffic-200) | OK | successfully updated service traffic |  | [schema](#update-namespace-service-traffic-200-schema) |

#### Responses


##### <span id="update-namespace-service-traffic-200"></span> 200 - successfully updated service traffic
Status: OK

###### <span id="update-namespace-service-traffic-200-schema"></span> Schema

###### Inlined models

**<span id="update-namespace-service-traffic-body"></span> UpdateNamespaceServiceTrafficBody**


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| values | [][UpdateNamespaceServiceTrafficParamsBodyValuesItems0](#update-namespace-service-traffic-params-body-values-items0)| `[]*UpdateNamespaceServiceTrafficParamsBodyValuesItems0` | ✓ | | List of revision traffic targets |  |



**<span id="update-namespace-service-traffic-params-body-values-items0"></span> UpdateNamespaceServiceTrafficParamsBodyValuesItems0**


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| percent | integer| `int64` |  | | Target traffice percentage |  |
| revision | string| `string` |  | | Target service revision |  |



### <span id="update-workflow"></span> Update a Workflow (*updateWorkflow*)

```
POST /api/namespaces/{namespace}/tree/{workflow}?op=update-workflow
```

Updates a workflow at the target path.
The body of this request should contain the workflow yaml you want to update to.


#### Consumes
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| workflow | `path` | string | `string` |  | ✓ |  | path to target workflow |
| workflow data | `body` | string | `string` | |  | | Payload that contains the updated direktiv workflow yaml. |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#update-workflow-200) | OK | successfully updated workflow |  | [schema](#update-workflow-200-schema) |

#### Responses


##### <span id="update-workflow-200"></span> 200 - successfully updated workflow
Status: OK

###### <span id="update-workflow-200-schema"></span> Schema

### <span id="watch-global-service-revision"></span> Watch Global Service Revision (*watchGlobalServiceRevision*)

```
GET /api/functions/{serviceName}/revisions/{revisionGeneration}
```

Watch a global scoped knative service revision.
The target revision generation is the number suffix on a revision.
Example: A revisions named 'global-fast-request-00003' would have the revisionGeneration '00003'.
Note: This is a Server-Sent-Event endpoint, and will not work with the default swagger client.


#### Produces
  * text/event-stream

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| revisionGeneration | `path` | string | `string` |  | ✓ |  | target revision generation |
| serviceName | `path` | string | `string` |  | ✓ |  | target service name |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#watch-global-service-revision-200) | OK | successfully watching service revision |  | [schema](#watch-global-service-revision-200-schema) |

#### Responses


##### <span id="watch-global-service-revision-200"></span> 200 - successfully watching service revision
Status: OK

###### <span id="watch-global-service-revision-200-schema"></span> Schema

### <span id="watch-global-service-revision-list"></span> Watch Global Service Revision List (*watchGlobalServiceRevisionList*)

```
GET /api/functions/{serviceName}/revisions
```

Watch the revision list of a global scoped knative service.
Note: This is a Server-Sent-Event endpoint, and will not work with the default swagger client.


#### Produces
  * text/event-stream

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| serviceName | `path` | string | `string` |  | ✓ |  | target service name |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#watch-global-service-revision-list-200) | OK | successfully watching service revisions |  | [schema](#watch-global-service-revision-list-200-schema) |

#### Responses


##### <span id="watch-global-service-revision-list-200"></span> 200 - successfully watching service revisions
Status: OK

###### <span id="watch-global-service-revision-list-200-schema"></span> Schema

### <span id="watch-namespace-service-revision"></span> Watch Namespace Service Revision (*watchNamespaceServiceRevision*)

```
GET /api/functions/namespaces/{namespace}/function/{serviceName}/revisions/{revisionGeneration}
```

Watch a namespace scoped knative service revision.
The target revision generation is the number suffix on a revision.
Example: A revisions named 'namespace-direktiv-fast-request-00003' would have the revisionGeneration '00003'.
Note: This is a Server-Sent-Event endpoint, and will not work with the default swagger client.


#### Produces
  * text/event-stream

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| revisionGeneration | `path` | string | `string` |  | ✓ |  | target revision generation |
| serviceName | `path` | string | `string` |  | ✓ |  | target service name |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#watch-namespace-service-revision-200) | OK | successfully watching service revision |  | [schema](#watch-namespace-service-revision-200-schema) |

#### Responses


##### <span id="watch-namespace-service-revision-200"></span> 200 - successfully watching service revision
Status: OK

###### <span id="watch-namespace-service-revision-200-schema"></span> Schema

### <span id="watch-namespace-service-revision-list"></span> Watch Namespace Service Revision List (*watchNamespaceServiceRevisionList*)

```
GET /api/functions/namespaces/{namespace}/function/{serviceName}/revisions
```

Watch the revision list of a namespace scoped knative service.
Note: This is a Server-Sent-Event endpoint, and will not work with the default swagger client.


#### Produces
  * text/event-stream

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| serviceName | `path` | string | `string` |  | ✓ |  | target service name |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#watch-namespace-service-revision-list-200) | OK | successfully watching service revisions |  | [schema](#watch-namespace-service-revision-list-200-schema) |

#### Responses


##### <span id="watch-namespace-service-revision-list-200"></span> 200 - successfully watching service revisions
Status: OK

###### <span id="watch-namespace-service-revision-list-200-schema"></span> Schema

### <span id="workflow-metrics-invoked"></span> Gets Invoked Workflow Metrics (*workflowMetricsInvoked*)

```
GET /api/namespaces/{namespace}/tree/{workflow}?op=metrics-invoked
```

Get metrics of invoked workflow instances.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| workflow | `path` | string | `string` |  | ✓ |  | path to target workflow |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#workflow-metrics-invoked-200) | OK | successfully got workflow metrics |  | [schema](#workflow-metrics-invoked-200-schema) |

#### Responses


##### <span id="workflow-metrics-invoked-200"></span> 200 - successfully got workflow metrics
Status: OK

###### <span id="workflow-metrics-invoked-200-schema"></span> Schema

### <span id="workflow-metrics-milliseconds"></span> Gets Workflow Time Metrics (*workflowMetricsMilliseconds*)

```
GET /api/namespaces/{namespace}/tree/{workflow}?op=metrics-failed
```

Get the timing metrics of a workflow's instance.
This returns a total sum of the milliseconds a workflow has been executed for.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| workflow | `path` | string | `string` |  | ✓ |  | path to target workflow |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#workflow-metrics-milliseconds-200) | OK | successfully got workflow metrics |  | [schema](#workflow-metrics-milliseconds-200-schema) |

#### Responses


##### <span id="workflow-metrics-milliseconds-200"></span> 200 - successfully got workflow metrics
Status: OK

###### <span id="workflow-metrics-milliseconds-200-schema"></span> Schema

### <span id="workflow-metrics-state-milliseconds"></span> Gets a Workflow State Time Metrics (*workflowMetricsStateMilliseconds*)

```
GET /api/namespaces/{namespace}/tree/{workflow}?op=metrics-state-milliseconds
```

Get the state timing metrics of a workflow's instance.
This returns the timing of individual states in a workflow.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| workflow | `path` | string | `string` |  | ✓ |  | path to target workflow |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#workflow-metrics-state-milliseconds-200) | OK | successfully got workflow metrics |  | [schema](#workflow-metrics-state-milliseconds-200-schema) |

#### Responses


##### <span id="workflow-metrics-state-milliseconds-200"></span> 200 - successfully got workflow metrics
Status: OK

###### <span id="workflow-metrics-state-milliseconds-200-schema"></span> Schema

### <span id="workflow-metrics-successful"></span> Gets Successful Workflow Metrics (*workflowMetricsSuccessful*)

```
GET /api/namespaces/{namespace}/tree/{workflow}?op=metrics-successful
```

Get metrics of a workflow, where the instance was successful.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| namespace | `path` | string | `string` |  | ✓ |  | target namespace |
| workflow | `path` | string | `string` |  | ✓ |  | path to target workflow |

#### All responses

| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#workflow-metrics-successful-200) | OK | successfully got workflow metrics |  | [schema](#workflow-metrics-successful-200-schema) |

#### Responses


##### <span id="workflow-metrics-successful-200"></span> 200 - successfully got workflow metrics
Status: OK

###### <span id="workflow-metrics-successful-200-schema"></span> Schema

## Models




# Direktiv API.
direktiv api
  

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

### Produces
  * application/json

## Access control

### Security Schemes

#### api_key (header: KEY)



> **Type**: apikey

### Security Requirements
  * api_key

## All endpoints

###  create_global_function

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| POST | /api/functions/{function} | [update service request](#update-service-request) | Creates a global function. |
  


###  operations

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/functions | [get functions1](#get-functions1) | Returns list of global functions. |
| GET | /api/functions/{function} | [get global functions](#get-global-functions) | Returns list of global functions. |
  


## Paths

### <span id="update-service-request"></span> Creates a global function. (*UpdateServiceRequest*)

```
POST /api/functions/{function}
```

Creates a global Knative function with 'global-' prefix.

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| function | `path` | string | `string` |  | ✓ |  | name of the function |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#update-service-request-200) | OK | service created |  | [schema](#update-service-request-200-schema) |

#### Responses


##### <span id="update-service-request-200"></span> 200 - service created
Status: OK

###### <span id="update-service-request-200-schema"></span> Schema

### <span id="get-functions1"></span> Returns list of global functions. (*getFunctions1*)

```
GET /api/functions
```

Returns list of global Knative functions with 'global-' prefix.

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [201](#get-functions1-201) | Created | service created |  | [schema](#get-functions1-201-schema) |

#### Responses


##### <span id="get-functions1-201"></span> 201 - service created
Status: Created

###### <span id="get-functions1-201-schema"></span> Schema

### <span id="get-global-functions"></span> Returns list of global functions. (*getGlobalFunctions*)

```
GET /api/functions/{function}
```

Returns list of global Knative functions with 'global-' prefix.

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| function | `path` | string | `string` |  | ✓ |  | name of the function |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-global-functions-200) | OK | service created |  | [schema](#get-global-functions-200-schema) |

#### Responses


##### <span id="get-global-functions-200"></span> 200 - service created
Status: OK

###### <span id="get-global-functions-200-schema"></span> Schema

## Models




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

###  operations

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/functions | [get functions](#get-functions) | Returns list of global functions. |
  


## Paths

### <span id="get-functions"></span> Returns list of global functions. (*getFunctions*)

```
GET /api/functions
```

Returns list of global Knative functions with 'global-' prefix.

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [201](#get-functions-201) | Created | service created |  | [schema](#get-functions-201-schema) |
| [500](#get-functions-500) | Internal Server Error | internal error |  | [schema](#get-functions-500-schema) |

#### Responses


##### <span id="get-functions-201"></span> 201 - service created
Status: Created

###### <span id="get-functions-201-schema"></span> Schema

##### <span id="get-functions-500"></span> 500 - internal error
Status: Internal Server Error

###### <span id="get-functions-500-schema"></span> Schema

## Models

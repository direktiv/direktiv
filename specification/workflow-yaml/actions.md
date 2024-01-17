# Actions

```yaml
- id: a
  type: action
  action:
    function: myfunc
    input: 'jq(.x)'
```

## ActionDefinition 

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `function` | Name of the referenced function. See [FunctionDefinition](functions.md#functiondefinition). | string | yes |
| `input` | Selects or generates the data to send as input to the function. | [Structured JQ](../instance-data/structured-jx.md) | no |
| `secrets` | Defines a list of secrets to temporarily add to the instance data under `.secrets`, before evaluating the `input`. | []string | no |
| `retries` | | [[]RetryPolicyDefinition](#retrypolicydefinition) | no |
| `files` | Determines a list of files to load onto the function's file-system from variables. Only valid if the referenced function supports it. | [[]FunctionFileDefinition](#functionfiledefinition) | no |

## RetryPolicyDefinition 

```yaml
- id: a
  type: action
  action:
    function: myfunc
    input: 'jq(.x)'
    retries:
      codes: [".*"]
      max_attempts: 3
      delay: PT3S
      multiplier: 1.5
```

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| codes | A list of "glob" patterns that will be compared to catchable error codes returned by the function to determine if this retry policy applies. | []string | yes |
| max_attempts | Maximum number of retry attempts. If the function has been retried this many times or more when this policy is invoked the retry will be skipped, and instead the error will be escalated to the state's error handling logic.  | integer | yes |
| delay | ISO8601 duration string giving a time delay between retry attempts. | string | no |
| multiplier | Value by which the delay is multiplied after each attempt. | float | no |

## FunctionFileDefinition

```yaml
- id: a
  type: action
  action:
    function: myfunc
    input: 'jq(.x)'
    files:
    - key: VAR_A 
      scope: namespace
      as: a
```

Some function types support loading variable directly from storage onto their file-systems. This object defines what variable to load and what to save it as.

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `key` | Identifies which variable to load into a file. | string | yes | 
| `scope` | Specifies the scope from which to load the variable. | [VariableScopeDefinition](#variablescopedefinition) | no |
| `as` | Names the resulting file. If left unspecified, the `key` will be used instead. | string | no |

## VariableScopeDefinition

Every variable exists within a single scope. The scope dictates what can access it and how persistent it is. There are four defined [scopes](../variables/variables.md):

* `instance`
* `workflow`
* `namespace`
* `file`

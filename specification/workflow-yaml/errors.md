# Errors

Errors can happen for many reasons. Direktiv allows you to catch and handle these errors using a common field 'catch'. This field takes an array of [ErrorCatchDefinition](#errorcatchdefinition) objects, each specifying one or more errors that apply and where to transition to next in order to handle them. When an error is thrown, the list of error catchers is evaluated in order until a match is found. If no match is found, the instance fails. 

```yaml
direktiv_api: workflow/v1
states:
- id: a
  type: consumeEvent
  timeout: PT5S
  event:
    type: com.github.pull.create
  catch: 
  - error: "direktiv.cancels.timeout.soft"
    transition: handle-error
- id: handle-error
  type: noop
  log: handling error
```

## ErrorCatchDefinition

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| error | Specified what error code(s) this catcher applies to. This should be a "glob" pattern that will be compared to catchable error codes to determine if this retry policy applies. | string | yes |
| transition | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, this state terminates the workflow. | string | no |


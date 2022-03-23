# Errors

Errors can happen for many reasons. Direktiv allows you to catch and handle these errors using a common field 'catch'. This field takes an array of [ErrorCatchDefinition](#ErrorCatchDefinition) objects, each specifying one or more errors that apply and where to transition to next in order to handle them. When an error is thrown, the list of error catchers is evaluated in order until a match is found. If no match is found, the instance fails. 

```yaml
- id: a
  type: delay
  duration: PT5M
  catch: 
  - error: "direktiv.cancels.timeout.soft"
    transition: b
```

## ErrorCatchDefinition

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| error | Specified what error code(s) this catcher applies to. This should be a "glob" pattern that will be compared to catchable error codes to determine if this retry policy applies. | string | yes |
| transition | Identifies which state to transition to next, referring to the next state's unique `id`. If undefined, this state terminates the workflow. | string | no |


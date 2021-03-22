# Check Credit Score

This example demonstrates the use of a `switch` state in an event-based workflow. The state waits for the arrival of a `checkcredit` event, and conditionally 'approves' or 'rejects' a hypothetical loan request based on data included in the `checkcredit` event using a state.

## check-credit Workflow YAML
```yaml
id: check-credit
start:
  type: event
  state: Check-Credit
  event:
    type: checkcredit
states:
- id: Check-Credit
  type: switch
  conditions:
  - condition: '.checkcredit.value > 500'
    transition: Approve-Loan
  defaultTransition: Reject-Loan
- id: Reject-Loan
  type: noop
  transform: '{ "msg": "You have been rejected for this loan" }'
- id: Approve-Loan
  type: noop
  transform: '{ "msg": "You have been approved for this loan" }'
```

## gen-credit Workflow YAML
```yaml
id: generate-credit
description: "Generate credit score event" 
states:
- id: gen
  type: generateEvent
  event:
    type: checkcredit
    source: Direktiv
    data: '{
      "value": 501
    }'
```

# Event Transition 

## Workflow

```yaml
id: eventtransition
states:
- id: ingress
  type: eventXor
  events:
  - event:
      type: visaApprovedEvent
      context:
        source: visaCheckSource
    transition: HandleApprovedVisa
  - event:
      type: visaRejectedEvent
      context:
        source: visaCheckSource
    transition: HandleRejectedVisa
  timeout: PT1H
  default: HandleNoVisaDecision
- id: HandleApprovedVisa
  type: action
  action:
    workflow: handleApprovedVisaDecisionWorkflow
- id: HandleRejectedVisa
  type: action
  action:
    workflow: handleRejectedVisaDecisionWorkflow
- id: HandleNoVisaDecision
  type: action
  action:
    workflow: handleNoVisaDecisionWorkflow
```
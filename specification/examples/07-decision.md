# Decision

## Workflow

```yaml
id: decision
functions:
- id: sendRejectionEmail
  type: reusable
  image: apps.vorteil.io/direktive-demos/reject-email
states:
- id: CheckApplication
  type: switch
  conditions:
  - condition: '.applicant.age >= 18'
    transition: StartApplication
  default: RejectApplication
- id: StartApplication
  type: action
  action:
    workflow: startApplicationWorkflow
    input: '.applicant'
- id: RejectApplication
  type: action
  action:
    function: sendRejectionEmail
    input: '.applicant'
```

### Input

```json
{
	"applicant": {
		"fname": "Trent",
		"lname": "Hilliam",
		"age": 27,
		"email": "trent.hilliam@vorteil.io"
	}
}
```


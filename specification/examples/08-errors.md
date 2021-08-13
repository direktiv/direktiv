# Errors 

## Workflow 

```yaml
id: errors
functions:
- id: provisionOrderFunction
  type: reusable
  image: apps.vorteil.io/direktive-demos/provision-order
states:
- id: ProvisionOrder
  type: action
  action:
    function: provisionOrderFunction
    input: '.order'
  transition: ApplyOrder
  catch:
  - error: provision.missingID
    transition: MissingID
  - error: provision.missingItem
    transition: MissingItem
  - error: provision.missingQuantity
    transition: MissingQuantity
- id: ApplyOrder
  type: action
  action:
    workflow: applyOrderWorkflow
- id: MissingID
  type: action
  action:
    workflow: handleMissingIDException
- id: MissingItem
  type: action
  action:
    workflow: handleMissingItemException
- id: MissingQuantity
  type: action
  action:
    workflow: handleMissingQuantityException
```

## Input

```json
{
	"order": {
		"id": "",
		"item": "laptop",
		"quantity": "10"
	}
}
```
# Workflow Definition 

# Direktiv Workflow Definition

This document describes the rules for Direktiv workflow definition files. These files are written in YAML and dictate the behaviour of a workflow running on Direktiv. 

```yaml title="Workflow"
direktiv_api: workflow/v1
description: |
  A simple "Hello, world" demonstration.
states:
- id: hello
  type: noop
  transform: 'jq({ msg: "Hello, world!" })'
```

```json title="Input"
{}
```

```json title="Output"
{
  "msg": "Hello, world!"
}
```

Workflows have inputs and outputs, usually in JSON. Where examples appear in this document they will often be accompanied by inputs and outputs as seen above.

### WorkflowDefinition

This is the top-level structure of a Direktiv workflow definition. All workflows must have one.

| Parameter | Description | Type | Required |
| --- | --- | --- | --- |
| `direktiv_api` | Set it to 'workflow/v1' to reduce ambiguity, enabling tools to better identify this file as a workflow. | string | no |
| `url` | Link to further information. | string | no |
| `description` | Short description of the workflow.  | string | no |
| `functions` | List of function definitions for use by function-based `states`. | [[]FunctionDefinition](functions.md#functiondefinition) | no |
| `start` | Configuration for how the workflow should start. | [StartDefinition](./starts.md) | no |
| `states` | List of all possible workflow states. | [[]StateDefinition](./states.md) | yes | 
| `timeouts` | Configuration of workflow-level timeouts. | [TimeoutsDefinition](./timeouts.md) | no |

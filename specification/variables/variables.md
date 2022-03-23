# Variables

Direktiv can store data separately to [instance data](../instance-data/instance-data.md). An instance can read and change its instance data at will so you might wonder why this separation needs to exist, but it turns out variables solve a number of problems: 

* Efficiently passing around large datasets or files to actions, especially ones that exceed instance data size limits. 
* Persisting data between instances of a workflow.
* Sharing data between different workflows.

## Scopes

All variables belong to a scope. The scopes are `instance`, `workflow`, and `namespace`. Instance scoped variables are only accessible to the singular instance that created them. Workflow scoped variables can be used and shared between multiple instances of the same workflow. Namespace scoped variables are available to all instances of all workflows on the namespace. All variables are identified by a name, and each name is unique within its scope. 

## States

Two types of states in the workflow spec interact directly with variables: [`getter`](../workflow-yaml/getter.md) and [`setter`](../workflow-yaml/getter.md).

## Files 

Due to size limitations on action inputs and instance data it can sometimes be impossible to pass data to actions without using variables. Actions can interact with variables directly, loading them onto their file-system and sometimes creating/changing variables as well. 


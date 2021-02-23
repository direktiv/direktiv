# Parallel

## Workflow

```yaml
id: parallelexec
states:
- id: ParallelExec
  type: parallel
  mode: and 
  actions:
  - workflow: shortdelayworkflowid
  - workflow: longdelayworkflowid
```


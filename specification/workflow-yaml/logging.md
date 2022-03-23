# Logging 

All states can write to instance logs via a common field `log`. This field uses [structured jx](../instance-data/structured-jx.md) to support querying instance data and inserting it into the logs. 

```yaml
- id: a
  type: noop
  log: 'Hello, world!'
```


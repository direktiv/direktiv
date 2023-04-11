# Instance Metadata 

Instance metadata is a way to monitor an instance. An instance can update its metadata at any time, replacing it with whatever information it needs to expose via the API.

All states can write to instance metadata via a common field `metadata`. This field uses [structured jx](../instance-data/structured-jx.md) to support querying instance data and inserting it into the metadata. 

```yaml
states:
- id: a
  type: delay
  duration: PT1M
  metadata: 
    workflow-data: jq(.)
```

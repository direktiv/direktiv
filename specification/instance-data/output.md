# Instance Output 

Unlike instance data, which is not accessible to anything other than the instance itself, instance output is exposed via API. It is also returned to caller instance if the workflow was invoked as a subflow. 

When an instance completes it saves its instance data as its output, which indirectly exposes the instance data. Normally this is the desired behaviour, but it can present a security risk if handled incorrectly. 

Workflows should take steps to trim things they don't mean to return before terminating using [transforms](./transforms.md).

## Large Outputs

Like instance data, output data has size limits. These size limits are usually the same, but not necessarily. This will vary according to the configuration of each Direktiv installation, and is usually about 32 MiB. 

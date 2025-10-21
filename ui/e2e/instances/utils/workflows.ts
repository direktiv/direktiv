export const workflowWithDelay = `direktiv_api: workflow/v1
description: A simple 'delay' state that waits for 2 seconds
states:
- id: delay
  type: delay
  duration: PT2S
  transform:
    result: finished
`;

export const workflowWithFewLogs = `direktiv_api: workflow/v1
description: A simple 'delay' state that waits for 4 seconds
states:
- id: delay
  type: delay
  duration: PT4S
  transition: logs
- id: logs
  type: noop
  log: hello-world
  transform:
    result: Hello world!
`;

export const workflowWithManyLogs = `direktiv_api: workflow/v1
description: Produces a fair amount of logs
states:
  - id: prep
    type: noop
    transform:
      x: 10
    transition: loop
  - id: loop
    type: switch
    conditions:
      - condition: "jq(.x > 0)"
        transition: subtract-one
  - id: subtract-one
    type: delay
    duration: PT1S
    transition: loop
    transform:
      x: "jq(.x - 1)"
    `;

export const workflowThatFails = `description: A simple workflow that throws an error'
states:
- id: error
  type: error
  error: i-am-an-error
  message: this is my error message
`;

export const parentWorkflow = ({
  childPath,
  children = 1,
}: {
  childPath: string;
  children?: number;
}) => `description: I will spawn multiple instances of the child.yaml
functions:
- id: get
  workflow: ${childPath}
  type: subflow
states:
- id: prep 
  type: noop 
  transform:
    x: ${children}  # how many child instances should we spawn
  transition: loop
- id: loop
  type: switch
  conditions:
  - condition: 'jq(.x > 0)'
    transition: getter
- id: getter 
  type: action
  action:
    function: get
    input: 
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/todos/1"
  transition: loop
  transform: 
    x: 'jq(.x - 1)'`;

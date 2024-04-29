export const simpleWorkflow = `description: A simple 'no-op' state that returns 'Hello world!'
states:
- id: helloworld
  type: noop
  transform:
    result: Hello world!
`;

export const workflowWithDelay = `direktiv_api: workflow/v1
description: A simple 'delay' state that waits for 8 seconds
states:
- id: delay
  type: delay
  duration: PT8S
  transform:
    result: finished
`;

export const workflowWithDelayBeforeLogging = `direktiv_api: workflow/v1
description: A simple 'delay' state that waits for 3 seconds
states:
- id: delay
  type: delay
  duration: PT6S
  transition: logs
- id: logs
  type: noop
  log: log-message
  transform:
    result: Hello world!
  `;

export const workflowWithManyLogs = `direktiv_api: workflow/v1
description: Produces a fair amount of logs
states:
  - id: prep
    type: noop
    transform:
      x: 15
    transition: loop
  - id: loop
    type: switch
    conditions:
      - condition: "jq(.x > 0)"
        transition: subtract-one
  - id: subtract-one
    type: noop
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

export const workflowThatWaitsAndFails = `direktiv_api: workflow/v1
states:
- id: delay
  type: delay
  duration: PT5S
  transition: handle-error
- id: handle-error
  type: error
  error: i-am-an-error
  message: error-message
  transform:
    result: an error occurred
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

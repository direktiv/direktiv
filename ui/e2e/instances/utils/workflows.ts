export const workflowWithFewLogs = `const flow: FlowDefinition = {
  type: "default",
  timeout: "PT1S",
  state: "stateFirst",
};

const error = 'input must contain { "data": "string" or number }'

function stateFirst(input): StateFunction<unknown> {
  const { data } = input;
  if (!data) {
    return finish({ error });
  }
  return transition(stateSecond, data);
}

function formatMessage(data: string | number, type: string) {
  return { message: formatted }
}`;

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

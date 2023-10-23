export const simpleWorkflow = `description: A simple 'no-op' state that returns 'Hello world!'
states:
- id: helloworld
  type: noop
  transform:
    result: Hello world!
`;

export const workflowThatFails = `description: A simple workflow that throws an error'
states:
- id: error
  type: error
  error: i-am-an-error
  message: this is my error message
`;

export const parentWorkflow = ({
  childName,
  children = 1,
}: {
  childName: string;
  children?: number;
}) => `description: I will spawn multiple instances of the child.yaml
functions:
- id: get
  workflow: ${childName}
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

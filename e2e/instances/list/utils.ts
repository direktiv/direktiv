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

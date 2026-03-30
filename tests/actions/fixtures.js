export const workflowUsingActionSrc = `
const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateFirst",
};

const bash = generateAction({
  image: "direktiv/bash:dev",
  size: "small",
  envs: [
    {
      name: "myenv",
      value: "myenvvalue",
    },
  ],
});

function stateFirst() {
  const payload = {
    commands: [
      {
        command: "printenv myenv",
      },
    ],
  };

  const result = bash(payload);
  return finish(result);
}
`

export const workflowMultiCommandActionSrc = `
const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateFirst",
};

const bash = generateAction({
  image: "direktiv/bash:dev",
  size: "small",
});

function stateFirst() {
  const payload = {
    commands: [
      {
        command: "printf '%s' proof-one",
      },
      {
        command: "printf '%s' proof-two",
      },
      {
        command: "printf '%s' proof-three",
      },
    ],
  };

  const result = bash(payload);
  return finish(result);
}
`

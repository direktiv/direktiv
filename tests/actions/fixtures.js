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

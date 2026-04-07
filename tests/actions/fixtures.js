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

  const result = bash({ payload });
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

  const result = bash({ payload });
  return finish(result);
}
`

export const workflowEchoActionSrc = `
const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateFirst",
};

const echo = generateAction({
  type: "workflow",
  image: "direktiv/echo:dev",
  size: "small",
});

function stateFirst(input: unknown) {
  const result = echo({ payload: input });
  return finish(result);
}
`

export const workflowTwoActionsSrc = `
const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateFirst",
};

const bash = generateAction({
  image: "direktiv/bash:dev",
  size: "small",
});

const echo = generateAction({
  type: "workflow",
  image: "direktiv/echo:dev",
  size: "small",
});

function stateFirst() {
  const bashResult = bash({ payload: { commands: [{ command: "echo hi" }] } });

  const echoResult = echo({ payload: { hi: "there" } });

  return finish({
    bash: bashResult,
    echo: echoResult,
  });
}
`

export const workflowFilesInActionSrc = `const flow: FlowDefinition = {
  type: "default",
  timeout: "PT5M",
  state: "stateRunAction",
};

const action = generateAction({
  image: "direktiv/bash:dev",
});

function stateRunAction() {
  const result = action({
    payload: {
      commands: [
        {
          command: "ls -la",
        },
        {
          command: "cat wf-var-one",
        },
        {
          command: "cat test.wf.ts",
        },
      ],
    },
    files: [
      { name: "wf-var-one", scope: "workflow" },
      { name: "/test.wf.ts", scope: "file", permission: "0600" },
    ],
  });

  return finish(result);
}
`

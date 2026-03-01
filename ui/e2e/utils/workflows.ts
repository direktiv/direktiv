import { createFile } from "./files";

export const simpleWorkflow = `// A simple 'no-op' state that returns 'Hello world!'
const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateHello",
};

function stateHello(): StateFunction<unknown> {
  return finish({ message: "Hello world!" })
};
`;

export const errorWorkflow = `// This workflow will fail unless provided the right input
const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateError",
};

function stateError(input): StateFunction<unknown> {
  if (input === "don't fail me") {
    return finish("ok")
  }
  throw new Error("this was set up to fail")
};
`;

export const selectStateThroughInputWorkflow = `// this will progress through state B or state C dependent on input
const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateA",
};

function stateA(input): StateFunction<unknown> {
  const { data } = input;
  if (data === "B") {
    return transition(stateB, {});
  }
  return transition(stateC, {});
}

function stateB(): StateFunction<unknown> {
  return finish({ message: "finishing from stateB" });
}

function stateC(): StateFunction<unknown> {
  return finish({ message: "finishing from stateC" });
}
`;

export const delayWorkflow1s = `// This workflow waits for a number of seconds. You can specify the length 
// in the workflow input. For example, { "time": 10 }.\

const flow: FlowDefinition = {
  type: "default",
  timeout: "PT1S",
  state: "stateDelay",
};

function stateDelay(): StateFunction<unknown> {
  sleep(1)
  return finish({
    message: "Hello world!" 
  })
};`;

export const delayWorkflow5s = `// This workflow waits for a number of seconds. You can specify the length 
// in the workflow input. For example, { "time": 10 }.\

const flow: FlowDefinition = {
  type: "default",
  timeout: "PT1S",
  state: "stateDelay",
};

function stateDelay(): StateFunction<unknown> {
  sleep(5)
  return finish({
    message: "Hello world!" 
  })
};`;

export const workflowWithService = `const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateFirst",
};

const d = generateAction({
  image: "direktiv/request:v4",
  size: "small",
  envs: [
    {
      name: "MY_ENV_VAR",
      value: "env-var-value",
    },
  ],
});

function stateFirst(): StateFunction<unknown> {
  var payload = {
    commands: [
      {
        command: "ls -la",
      },
    ],
  };
  let result = d(payload);
  return finish(result);
}
`;

export const createWorkflow = async (namespace: string, name: string) => {
  const response = await createFile({
    namespace,
    name,
    type: "workflow",
    content: simpleWorkflow,
    mimeType: "application/x-typescript",
  });

  if (response.data.type !== "workflow") {
    throw "unexpected response when creating test file";
  }
  return name;
};

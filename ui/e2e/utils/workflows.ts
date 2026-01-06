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

export const delayWorkflow = `// This workflow waits for a number of seconds. You can specify the length 
// in the workflow input. For example, { "time": 10 }.\

const flow: FlowDefinition = {
  type: "default",
  timeout: "PT1S",
  state: "stateDelay",
};

function stateDelay(): StateFunction<unknown> {
  let seconds = 1;
  sleep(seconds)
  return finish({
    message: "Hello world!" 
  })
};`;

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

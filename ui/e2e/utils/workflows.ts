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
  timeout: "PT30S",
  state: "stateDelay",
};

function stateDelay(input): StateFunction<unknown> {
  let seconds = 2;
  if (typeof input === "object" && typeof input.time === "number") {
    seconds = input.time
  }
  sleep(seconds)
  return finish({
    "message": \`waited for \${seconds}s.\`
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

import { createFile } from "./files";

const noopYaml = `// A simple 'no-op' state that returns 'Hello world!
const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateHello",
};

function stateHello(): StateFunction<unknown> {
  return finish("Hello world!")
};
`;

export const createWorkflow = async (namespace: string, name: string) => {
  const response = await createFile({
    namespace,
    name,
    type: "workflow",
    content: noopYaml,
    mimeType: "application/x-typescript",
  });

  if (response.data.type !== "workflow") {
    throw "unexpected response when creating test file";
  }
  return name;
};

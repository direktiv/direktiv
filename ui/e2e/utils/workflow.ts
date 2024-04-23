import { createFile } from "./files";

export const noopYaml = `\
description: A simple 'no-op' state that returns 'Hello world!'
states:
- id: helloworld
  type: noop
  transform:
    result: Hello world!
`;

export const createWorkflow = async (namespace: string, name: string) => {
  const response = await createFile({
    namespace,
    name,
    type: "workflow",
    yaml: noopYaml,
  });

  if (response.data.type !== "workflow") {
    throw "unexpected response when creating test file";
  }
  return name;
};

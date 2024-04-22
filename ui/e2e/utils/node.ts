import { getFile as apiGetFile } from "~/api/files/query/file";
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

type ErrorType = { json?: { code?: string } };

export const checkIfFileExists = async ({
  namespace,
  path,
}: {
  namespace: string;
  path: string;
}) => {
  try {
    const response = await apiGetFile({
      urlParams: {
        baseUrl: process.env.PLAYWRIGHT_UI_BASE_URL,
        path,
        namespace,
      },
    });

    if (!response.data) {
      throw `Fetching file at ${path} in namespace ${namespace} failed`;
    }

    return response.data.path === path;
  } catch (error) {
    const typedError = error as ErrorType;

    if (typedError?.json?.code === "resource_not_found") {
      return false;
    }

    throw new Error(
      `Unexpected error fetching ${path} in namespace ${namespace}`
    );
  }
};

import { CreateFileSchemaType } from "~/api/files/schema";
import { createFile as apiCreateFile } from "~/api/files/mutate/createFile";
import { getFile as apiGetFile } from "~/api/files/query/file";
import { encode } from "js-base64";
import { headers } from "./testutils";

export const createFile = async ({
  name,
  yaml,
  namespace,
  type,
  path = "/",
}: {
  name: string;
  yaml: string;
  namespace: string;
  type: CreateFileSchemaType["type"];
  path?: string;
}) =>
  await apiCreateFile({
    payload: {
      data: encode(yaml),
      name,
      mimeType: "application/yaml",
      type,
    },
    urlParams: {
      baseUrl: process.env.PLAYWRIGHT_UI_BASE_URL,
      namespace,
      path,
    },
    headers,
  });

export const createDirectory = async ({
  name,
  namespace,
  path = "/",
}: {
  name: string;
  namespace: string;
  path?: string;
}) =>
  await apiCreateFile({
    payload: {
      name,
      type: "directory",
    },
    urlParams: {
      baseUrl: process.env.PLAYWRIGHT_UI_BASE_URL,
      namespace,
      path,
    },
    headers,
  });

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

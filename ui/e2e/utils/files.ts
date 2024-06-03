import { CreateFileSchemaType } from "~/api/files/schema";
import { createFile as apiCreateFile } from "~/api/files/mutate/createFile";
import { deleteFile as apiDeleteFile } from "~/api/files/mutate/deleteFile";
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

export const deleteFile = async ({
  namespace,
  path,
}: {
  namespace: string;
  path: string;
}) => {
  await apiDeleteFile({
    urlParams: {
      namespace,
      baseUrl: process.env.PLAYWRIGHT_UI_BASE_URL,
      path,
    },
    headers,
  });
};

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

type ErrorType = { status?: number };

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
      headers,
    });

    if (!response.data) {
      throw `Fetching file at ${path} in namespace ${namespace} failed`;
    }

    return response.data.path === path;
  } catch (error) {
    const typedError = error as ErrorType;
    if (typedError?.status === 404) {
      // fail silently to allow for using poll() in tests
      return false;
    }

    throw new Error(
      `Unexpected error ${typedError?.status} fetching ${path} in namespace ${namespace}`
    );
  }
};

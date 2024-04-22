import { CreateFileSchemaType } from "~/api/files/schema";
import { createFile as apiCreateFile } from "~/api/files/mutate/createFile";
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

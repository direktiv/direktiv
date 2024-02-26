import { FileTypeType } from "~/api/files/schema";
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
  type: FileTypeType;
  path?: string;
}) => {
  await apiCreateFile({
    payload: {
      data: encode(yaml),
      name,
      mimeType: "application/yaml",
      type,
    },
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      path,
    },
    headers,
  });
};

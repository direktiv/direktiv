import { createFile as apiCreateFile } from "~/api/files/mutate/createFile";
import { encode } from "js-base64";
import { headers } from "../../utils/testutils";

type CreateGatewayFileParams = {
  fileContent: string;
  namespace: string;
  name: string;
};

export const createGatewayFile = async ({
  fileContent,
  namespace,
  name,
}: CreateGatewayFileParams): Promise<void> => {
  await apiCreateFile({
    urlParams: {
      namespace,

      baseUrl: process.env.PLAYWRIGHT_UI_BASE_URL,
    },
    payload: {
      name,
      type: "gateway",
      mimeType: "application/yaml",
      data: encode(fileContent),
    },
    headers,
  });
};

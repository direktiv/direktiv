import { createInstance as createInstanceRequest } from "~/api/instances/mutate/create";
import { headers } from "e2e/utils/testutils";

export const createInstance = async ({
  namespace,
  path,
  payload,
}: {
  namespace: string;
  path: string;
  payload?: string;
}) =>
  await createInstanceRequest({
    urlParams: {
      baseUrl: process.env.PLAYWRIGHT_UI_BASE_URL,
      namespace,
      path,
    },
    headers,
    payload,
  });

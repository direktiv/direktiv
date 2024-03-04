import { headers } from "e2e/utils/testutils";
import { runWorkflow } from "~/api/tree/mutate/runWorkflow";

export const createInstance = async ({
  namespace,
  path,
}: {
  namespace: string;
  path: string;
}) =>
  await runWorkflow({
    urlParams: {
      baseUrl: process.env.VITE_E2E_UI_DOMAIN,
      namespace,
      path,
    },
    headers,
  });

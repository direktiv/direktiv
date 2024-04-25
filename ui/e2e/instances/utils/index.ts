import { headers } from "e2e/utils/testutils";
import { runWorkflow } from "~/api/tree_obsolete/mutate/runWorkflow";

export const createInstance = async ({
  namespace,
  path,
  payload,
}: {
  namespace: string;
  path: string;
  payload?: string;
}) =>
  await runWorkflow({
    urlParams: {
      baseUrl: process.env.PLAYWRIGHT_UI_BASE_URL,
      namespace,
      path,
    },
    headers,
    payload,
  });

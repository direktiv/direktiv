import { createRevision } from "~/api/tree/mutate/createRevision";
import { createWorkflow } from "~/api/tree/mutate/createWorkflow";
import { headers } from "./testutils";
import { updateWorkflow } from "~/api/tree/mutate/updateWorkflow";

const getRevisionContentVariation = (revision: number) => `\
description: minimal workflow
states:
- id: revision-${revision}
  type: noop
`;

export const createWorkflowWithThreeRevisions = async (
  namespace: string,
  workflowName: string,
  path?: string
) => {
  const contentRevision1 = getRevisionContentVariation(1);
  const contentRevision2 = getRevisionContentVariation(2);
  const contentRevision3 = getRevisionContentVariation(3);

  const commonUrlParams = {
    baseUrl: process.env.VITE_DEV_API_DOMAIN,
    namespace,
    path: `${path ?? ""}${workflowName}`,
    headers,
  };

  // revision 1
  await createWorkflow({
    payload: contentRevision1,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      path,
      name: workflowName,
    },
    headers,
  });

  const firstRevision = await createRevision({
    urlParams: commonUrlParams,
    headers,
  });

  // revision 2
  await updateWorkflow({
    payload: contentRevision2,
    urlParams: commonUrlParams,
    headers,
  });
  const secondRevision = await createRevision({
    urlParams: commonUrlParams,
    headers,
  });

  // revision 3
  await updateWorkflow({
    payload: contentRevision3,
    urlParams: commonUrlParams,
    headers,
  });

  const thirdRevision = await createRevision({
    urlParams: commonUrlParams,
    headers,
  });

  return {
    workflowName,
    revisionsPayload: [
      contentRevision1,
      contentRevision2,
      contentRevision3,
    ] as const,
    revisionsReponse: [firstRevision, secondRevision, thirdRevision] as const,
  };
};

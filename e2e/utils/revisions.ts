import { createRevision } from "~/api/tree/mutate/createRevision";
import { createWorkflow } from "~/api/tree/mutate/createWorkflow";
import { updateWorkflow } from "~/api/tree/mutate/updateWorkflow";

const changeContentForRevisions = (revision: number) => `
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
  const contentRevision1 = changeContentForRevisions(1);
  const contentRevision2 = changeContentForRevisions(2);
  const contentRevision3 = changeContentForRevisions(3);

  const commonUrlParams = {
    baseUrl: process.env.VITE_DEV_API_DOMAIN,
    namespace,
    path: `${path ?? ""}${workflowName}`,
  };

  // revision 1
  await createWorkflow({
    payload: contentRevision1,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      path: path,
      name: workflowName,
    },
  });

  const firstRevision = await createRevision({
    payload: undefined,
    urlParams: commonUrlParams,
  });

  // revision 2
  await updateWorkflow({
    payload: contentRevision2,
    urlParams: commonUrlParams,
  });
  const secondRevision = await createRevision({
    payload: undefined,
    urlParams: commonUrlParams,
  });

  // revision 3
  await updateWorkflow({
    payload: contentRevision3,
    urlParams: commonUrlParams,
  });
  const thridRevision = await createRevision({
    payload: undefined,
    urlParams: commonUrlParams,
  });

  return {
    workflowName,
    revisionsPayload: [
      contentRevision1,
      contentRevision2,
      contentRevision3,
    ] as const,
    revisionsReponse: [firstRevision, secondRevision, thridRevision] as const,
  };
};

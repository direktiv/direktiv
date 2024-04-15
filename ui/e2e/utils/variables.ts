import { EditorMimeTypeSchema } from "~/components/VariableForm/utils";
import { createVar } from "~/api/variables/mutate/create";
import { encode } from "js-base64";
import { faker } from "@faker-js/faker";
import { forceLeadingSlash } from "~/api/files/utils";
import { headers } from "./testutils";

// Note: This makes sure only mimeTypes supported by the form are used,
// but the generated content isn't really in that format.
const { options: supportedMimeTypes } = EditorMimeTypeSchema;

export const createVariables = async (namespace: string, amount = 5) => {
  const variables = Array.from({ length: amount }, () => ({
    name: faker.internet.domainWord(),
    content: encode(faker.lorem.sentence()),
    mimeType:
      supportedMimeTypes[Math.floor(Math.random() * supportedMimeTypes.length)],
  }));

  return await Promise.all(
    variables.map((variable) =>
      createVar({
        payload: {
          name: variable.name,
          mimeType: variable.mimeType,
          data: variable.content,
        },
        urlParams: {
          baseUrl: process.env.PLAYWRIGHT_UI_BASE_URL,
          namespace,
        },
        headers: {
          ...headers,
        },
      }).then((result) => ({
        ...result,
        content: variable.content,
      }))
    )
  );
};

export const createWorkflowVariables = async (
  namespace: string,
  workflow: string,
  amount = 5
) => {
  // It may be advisable to keep "content" short so it is easier to test in the
  // editor (where each line is a separate HTML element)
  const variables = Array.from({ length: amount }, () => ({
    name: faker.internet.domainWord(),
    content: encode(faker.git.shortSha()),
    mimeType:
      supportedMimeTypes[Math.floor(Math.random() * supportedMimeTypes.length)],
  }));

  return await Promise.all(
    variables.map((variable) =>
      createVar({
        payload: {
          name: variable.name,
          mimeType: variable.mimeType,
          data: variable.content,
          workflowPath: forceLeadingSlash(workflow),
        },
        urlParams: {
          baseUrl: process.env.PLAYWRIGHT_UI_BASE_URL,
          namespace,
        },
        headers: {
          ...headers,
        },
      })
    )
  );
};

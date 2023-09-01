import { MimeTypeSchema } from "~/pages/namespace/Settings/Variables/MimeTypeSelect";
import { faker } from "@faker-js/faker";
import { headers } from "./testutils";
import { setVariable } from "~/api/tree/mutate/setVariable";

// Note: This makes sure only mimeTypes supported by the form are used,
// but the generated content isn't really in that format.
const { options } = MimeTypeSchema;

export const createWorkflowVariables = async (
  namespace: string,
  workflow: string,
  amount = 5
) => {
  const variables = Array.from({ length: amount }, () => ({
    name: faker.internet.domainWord(),
    content: faker.lorem.sentence(),
    mimeType: options[Math.floor(Math.random() * options.length)],
  }));

  for (let i = 0; i < amount; i++) {
    const variable = variables[i];
    await setVariable({
      payload: variable?.content,
      urlParams: {
        baseUrl: process.env.VITE_DEV_API_DOMAIN,
        namespace,
        path: workflow,
        name: variable?.name || "",
      },
      headers: {
        ...headers,
        "content-type": variable?.mimeType,
      },
    });
  }
};

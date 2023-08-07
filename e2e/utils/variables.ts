import { MimeTypeSchema } from "~/pages/namespace/Settings/Variables/MimeTypeSelect";
import { faker } from "@faker-js/faker";
import { headers } from "./testutils";
import { updateVar } from "~/api/variables/mutate/updateVariable";

// Note: This makes sure only mimeTypes supported by the form are used,
// but the generated content isn't really in that format.
const { options } = MimeTypeSchema;

export const createVariables = async (namespace: string, amount = 5) => {
  const variables = Array.from({ length: amount }, () => ({
    name: faker.internet.domainWord(),
    content: faker.lorem.sentence(),
    mimeType: options[Math.floor(Math.random() * options.length)],
  }));

  return await Promise.all(
    variables.map((variable) =>
      updateVar({
        payload: variable.content,
        urlParams: {
          baseUrl: process.env.VITE_DEV_API_DOMAIN,
          namespace,
          name: variable.name,
        },
        headers: {
          ...headers,
          "content-type": variable.mimeType,
        },
      }).then((result) => ({
        ...result,
        content: variable.content,
      }))
    )
  );
};

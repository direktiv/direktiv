import { createSecret } from "~/api/secrets/mutate/create";
import { encode } from "js-base64";
import { faker } from "@faker-js/faker";
import { headers } from "./testutils";

export const createSecrets = async (namespace: string, amount = 5) => {
  const secrets = Array.from({ length: amount }, () => ({
    name: faker.internet.domainWord(),
    value: encode(faker.internet.password()),
  }));

  return await Promise.all(
    secrets.map((secret) =>
      createSecret({
        payload: {
          name: secret.name,
          data: secret.value,
        },
        urlParams: {
          baseUrl: process.env.PLAYWRIGHT_UI_BASE_URL,
          namespace,
        },
        headers,
      })
    )
  );
};

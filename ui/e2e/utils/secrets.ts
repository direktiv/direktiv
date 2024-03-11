import { faker } from "@faker-js/faker";
import { headers } from "./testutils";
import { updateSecret } from "~/api/secrets/mutate/updateSecret";

export const createSecrets = async (namespace: string, amount = 5) => {
  const secrets = Array.from({ length: amount }, () => ({
    name: faker.internet.domainWord(),
    value: faker.internet.password(),
  }));

  return await Promise.all(
    secrets.map((secret) =>
      updateSecret({
        payload: secret.value,
        urlParams: {
          baseUrl: process.env.PLAYWRIGHT_UI_BASE_URL,
          namespace,
          name: secret.name,
        },
        headers,
      })
    )
  );
};

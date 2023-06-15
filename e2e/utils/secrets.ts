import { createSecret } from "~/api/secrets/mutate/createSecret";
import { faker } from "@faker-js/faker";

export const createSecrets = async (namespace: string, amount = 5) => {
  const secrets = Array.from({ length: amount }, () => ({
    name: faker.internet.domainWord(),
    value: faker.internet.password(),
  }));

  return await Promise.all(
    secrets.map((secret) =>
      createSecret({
        payload: secret.value,
        urlParams: {
          baseUrl: process.env.VITE_DEV_API_DOMAIN,
          namespace,
          name: secret.name,
        },
        headers: undefined,
      })
    )
  );
};

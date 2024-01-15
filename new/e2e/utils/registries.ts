import { createRegistry } from "~/api/registries/mutate/createRegistry";
import { faker } from "@faker-js/faker";
import { headers } from "./testutils";

export const createRegistries = async (namespace: string, amount = 5) => {
  const registries = Array.from({ length: amount }, () => ({
    url: faker.internet.url(),
    user: `${faker.internet.userName()}`,
    password: `${faker.internet.password()}`,
  }));

  return await Promise.all(
    registries.map((registry) =>
      createRegistry({
        payload: {
          user: registry.user,
          password: registry.password,
          url: registry.url,
        },
        urlParams: {
          baseUrl: process.env.VITE_DEV_API_DOMAIN,
          namespace,
        },
        headers,
      }).then(() => registry)
    )
  );
};

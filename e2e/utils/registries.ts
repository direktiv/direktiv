import { createRegistry } from "~/api/registries/mutate/createRegistry";
import { faker } from "@faker-js/faker";

export const createRegistries = async (namespace: string, amount = 5) => {
  const registries = Array.from({ length: amount }, () => ({
    data: `${faker.internet.userName()}:${faker.internet.password()}`,
    reg: faker.internet.url(),
  }));

  return await Promise.all(
    registries.map((registry) =>
      createRegistry({
        payload: registry,
        urlParams: {
          baseUrl: process.env.VITE_DEV_API_DOMAIN,
          namespace,
        },
        headers: undefined,
      })
    )
  );
};

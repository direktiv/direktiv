import { faker } from "@faker-js/faker";

const apiUrl = process.env.VITE_DEV_API_DOMAIN;

export const createNamespaceName = () => `playwright-${faker.git.shortSha()}`;

export const createNamespace = () =>
  new Promise<string>((resolve, reject) => {
    const name = createNamespaceName();
    fetch(`${apiUrl}/api/namespaces/${name}`, {
      method: "PUT",
    }).then(() => resolve(name));
  });

export const deleteNamespace = (namespace) =>
  new Promise<void>((resolve, reject) => {
    fetch(`${apiUrl}/api/namespaces/${namespace}?recursive=true`, {
      method: "DELETE",
    }).then(() => resolve());
  });

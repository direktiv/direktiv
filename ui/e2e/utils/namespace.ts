import { NamespaceListSchemaType } from "~/api/namespaces/schema/namespace";
import { faker } from "@faker-js/faker";
import { headers } from "./testutils";

const apiUrl = process.env.PLAYWRIGHT_UI_BASE_URL;

export const createNamespaceName = () => `playwright-${faker.git.shortSha()}`;

export const createNamespace = (name: string = createNamespaceName()) =>
  new Promise<string>((resolve, reject) => {
    fetch(`${apiUrl}/api/v2/namespaces/`, {
      method: "POST",
      headers,
      body: JSON.stringify({ name }),
    }).then((response) => {
      response.ok
        ? resolve(name)
        : reject(`creating namespace failed with code ${response.status}`);
    });
  });

export const deleteNamespace = (namespace: string) =>
  new Promise<void>((resolve, reject) => {
    fetch(`${apiUrl}/api/v2/namespaces/${namespace}`, {
      method: "DELETE",
      headers,
    }).then((response) => {
      response.ok
        ? resolve()
        : reject(`deleting namespace failed with code ${response.status}`);
    });
  });

export const checkIfNamespaceExists = async (namespace: string) => {
  const response = await fetch(`${apiUrl}/api/v2/namespaces`, { headers });
  if (!response.ok) {
    throw `fetching namespaces failed with code ${response.status}`;
  }
  const namespaceInResponse = await response
    .json()
    .then((json: NamespaceListSchemaType) =>
      json.data.find((ns) => ns.name === namespace)
    );
  return !!namespaceInResponse;
};

// Not intended for regular use. Namespaces should be cleaned up after every test.
// E.g., see the beforeEach and afterEach implementation in e2e/explorer.spec.ts.
// If you have spammed namespaces while writing tests, call this temporarily:
// await cleanupNamespace();
export const cleanupNamespaces = async () => {
  const response = await fetch(`${apiUrl}/api/v2/namespaces`, { headers });
  const namespaces = await response
    .json()
    .then((json: NamespaceListSchemaType) =>
      json.data.filter((ns) => ns.name.includes("playwright"))
    );
  const requests = namespaces.map((ns) =>
    fetch(`${apiUrl}/api/v2/namespaces/${ns.name}`, {
      method: "DELETE",
    })
  );

  return Promise.all(requests);
};

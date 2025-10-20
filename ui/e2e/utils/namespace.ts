import { NamespaceListSchemaType } from "~/api/namespaces/schema/namespace";
import { faker } from "@faker-js/faker";
import { headers } from "./testutils";

const apiUrl = process.env.PLAYWRIGHT_UI_BASE_URL;

export const createNamespaceName = () =>
  `playwright-${faker.git.commitSha({ length: 7 })}`;

export const createNamespace = (name: string = createNamespaceName()) =>
  new Promise<string>((resolve, reject) => {
    fetch(`${apiUrl}/api/v2/namespaces/`, {
      method: "POST",
      headers,
      body: JSON.stringify({ name }),
    }).then((response) => {
      if (response.ok) {
        resolve(name);
      } else {
        reject(`creating namespace failed with code ${response.status}`);
      }
    });
  });

export const deleteNamespace = (namespace: string) =>
  new Promise<void>((resolve, reject) => {
    fetch(`${apiUrl}/api/v2/namespaces/${namespace}`, {
      method: "DELETE",
      headers,
    }).then((response) => {
      if (response.ok) {
        resolve();
      } else {
        reject(`deleting namespace failed with code ${response.status}`);
      }
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

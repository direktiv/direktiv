import { NodeListSchemaType } from "~/api/tree/schema";
import { headers } from "./testutils";

const apiUrl = process.env.VITE_DEV_API_DOMAIN;

export const workflowExamples = {
  noop: `\
description: A simple 'no-op' state that returns 'Hello world!'
states:
- id: helloworld
  type: noop
  transform:
    result: Hello world!
`,
};

export const createWorkflow = (namespace: string, name: string) =>
  fetch(
    `${apiUrl}/api/namespaces/${namespace}/tree/${name}?op=create-workflow`,
    {
      method: "PUT",
      body: workflowExamples.noop,
      headers,
    }
  ).then((response) => {
    if (!response.ok) {
      throw `creating node failed with code ${response.status}`;
    }
    return name;
  });

export const createDirectory = (namespace: string, name: string) =>
  fetch(
    `${apiUrl}/api/namespaces/${namespace}/tree/${name}?op=create-directory`,
    {
      method: "PUT",
      headers,
    }
  ).then((response) => {
    if (!response.ok) {
      throw `creating node failed with code ${response.status}`;
    }
    return name;
  });

export const deleteNode = (namespace: string, name: string) =>
  fetch(`${apiUrl}/api/namespaces/${namespace}/tree/${name}?op=delete-node`, {
    method: "DELETE",
    headers,
  }).then((response) => {
    if (!response.ok) {
      throw `deleting node failed with code ${response.status}`;
    }
    return;
  });

export const checkIfNodeExists = (namespace: string, nodeName: string) =>
  fetch(`${apiUrl}/api/namespaces/${namespace}/tree`, { headers }).then(
    (response) => {
      if (!response.ok) {
        throw `fetching nodes failed with code ${response.status}`;
      }
      return response.json().then((json: NodeListSchemaType) => {
        const nodeInResponse = json?.children?.results
          .map((node) => node.name)
          .find((name) => name === nodeName);
        return !!nodeInResponse;
      });
    }
  );

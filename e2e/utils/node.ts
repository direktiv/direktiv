const apiUrl = process.env.VITE_DEV_API_DOMAIN;

export const workflowExamples = {
  noop: `
  description: A simple 'no-op' state that returns 'Hello world!'
  states:
  - id: helloworld
    type: noop
    transform:
      result: Hello world!
  `,
};

export const createWorkflow = (namespace: string, name: string) =>
  new Promise<string>((resolve, reject) => {
    fetch(
      `${apiUrl}/api/namespaces/${namespace}/tree/${name}?op=create-workflow`,
      {
        method: "PUT",
        body: workflowExamples.noop,
      }
    ).then((response) => {
      response.ok
        ? resolve(name)
        : reject(`creating node failed with code ${response.status}`);
    });
  });

export const createDirectory = (namespace: string, name: string) =>
  new Promise<string>((resolve, reject) => {
    fetch(
      `${apiUrl}/api/namespaces/${namespace}/tree/${name}?op=create-directory`,
      {
        method: "PUT",
      }
    ).then((response) => {
      response.ok
        ? resolve(name)
        : reject(`creating node failed with code ${response.status}`);
    });
  });

export const deleteNode = (namespace: string, type: Node, name: string) =>
  new Promise<void>((resolve, reject) => {
    fetch(
      `${apiUrl}/api/namespaces/${namespace}/tree/${name}?op=delete-${type}`,
      {
        method: "DELETE",
      }
    ).then((response) => {
      response.ok
        ? resolve()
        : reject(`deleting node failed with code ${response.status}`);
    });
  });

export const checkIfNodeExists = (namespace: string, nodeName: string) =>
  new Promise<boolean>((resolve, reject) => {
    fetch(`${apiUrl}/api/namespaces/${namespace}/tree`).then((response) => {
      if (!response.ok) {
        return reject(`fetching nodes failed with code ${response.status}`);
      }
      response.json().then((json) => {
        const nodeInResponse = json.children.results
          .map((node) => node.name)
          .find((name) => name === nodeName);
        resolve(!!nodeInResponse);
      });
    });
  });

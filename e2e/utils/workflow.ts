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

export const checkIfNodeExists = async (namespace, nodeName) => {
  const response = await fetch(`${apiUrl}/api/namespaces/${namespace}/tree`);
  const nodeInResponse = await response
    .json()
    .then((json) =>
      json.children.results
        .map((node) => node.name)
        .find((name) => name === nodeName)
    );
  return !!nodeInResponse;
};

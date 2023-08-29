export const generateApiCommandTemplate = (
  namespace: string,
  workflow: string
) => {
  const baseUrl = window.location.origin;
  return [
    {
      key: "execute",
      method: "POST",
      url: `${baseUrl}/api/namespaces/${namespace}/tree/${workflow}?op=execute`,
      body: `{}`,
      type: "json",
    },
    {
      key: "awaitExecute",
      method: "POST",
      url: `${baseUrl}/api/namespaces/${namespace}/tree/${workflow}?op=wait`,
      body: `{}`,
      type: "json",
    },
    {
      key: "update",
      method: "POST",
      url: `${baseUrl}/api/namespaces/${namespace}/tree/${workflow}?op=update-workflow`,
      body: `description: A simple 'no-op' state that returns 'Hello world!'
states:
- id: helloworld
type: noop
transform:
    result: Hello world!`,
      type: "yaml",
    },
  ] as const;
};

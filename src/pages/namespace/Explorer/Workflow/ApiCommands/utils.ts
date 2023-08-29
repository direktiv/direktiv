import { useMemo } from "react";
const baseUrl = window.location.origin;

export const useApiCommandTemplate = (namespace: string, workflow: string) => {
  const memoizedTemplates = useMemo(
    () =>
      [
        {
          key: "execute",
          method: "POST",
          url: `${baseUrl}/api/namespaces/${namespace}/tree/${workflow}?op=execute`,
          body: `{}`,
          payloadSyntax: "json",
        },
        {
          key: "awaitExecute",
          method: "POST",
          url: `${baseUrl}/api/namespaces/${namespace}/tree/${workflow}?op=wait`,
          body: `{}`,
          payloadSyntax: "json",
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
          payloadSyntax: "yaml",
        },
      ] as const,
    [namespace, workflow]
  );

  return memoizedTemplates;
};

export const useCurlCommand = ({
  url,
  body,
}: {
  url: string;
  body: string;
}) => `curl '${url}' \\
  -H 'direktiv-token: XXXXXXXXXXXXXXX' \\
  --data-raw $'${body}'`;

import { useApiKey } from "~/util/store/apiKey";
import useApiKeyHandling from "~/hooksNext/useApiKeyHandling";
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
          payloadSyntax: "json",
          body: `{
  "some": "input"
}`,
        },
        {
          key: "awaitExecute",
          method: "POST",
          url: `${baseUrl}/api/namespaces/${namespace}/tree/${workflow}?op=wait`,
          payloadSyntax: "json",
          body: `{
  "some": "input"
}`,
        },
        {
          key: "update",
          method: "POST",
          url: `${baseUrl}/api/namespaces/${namespace}/tree/${workflow}?op=update-workflow`,
          payloadSyntax: "yaml",
          body: `description: A simple 'no-op' state that returns 'Hello world!'
states:
- id: helloworld
  type: noop
  transform:
    result: Hello world!`,
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
}) => {
  const { isKeyRequired } = useApiKeyHandling();
  const apiKeyFromLocalstorage = useApiKey();

  const apiKeyHeader =
    isKeyRequired && apiKeyFromLocalstorage
      ? `\n -H 'direktiv-token: ${apiKeyFromLocalstorage}' \\`
      : "";

  const bodyEscaped = body.replace(/'/g, "\\'");

  return `curl '${url}' \\${apiKeyHeader}
  --data-raw $'${bodyEscaped}'`;
};

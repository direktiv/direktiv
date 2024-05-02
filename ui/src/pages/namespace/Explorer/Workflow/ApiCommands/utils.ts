import { getAuthHeader } from "~/api/utils";
import { useApiKey } from "~/util/store/apiKey";
import useApiKeyHandling from "~/hooks/useApiKeyHandling";
import { useMemo } from "react";

export const useApiCommandTemplate = (namespace: string, workflow: string) => {
  const baseUrl = window.location.origin;
  const memoizedTemplates = useMemo(
    () =>
      [
        {
          key: "execute",
          method: "POST",
          url: `${baseUrl}/api/v2/namespaces/${namespace}/instances?path=${workflow}`,
          payloadSyntax: "json",
          body: `{
  "some": "input"
}`,
        },
        {
          key: "awaitExecute",
          method: "POST",
          url: `${baseUrl}/api/v2/namespaces/${namespace}/instances?path=${workflow}&wait=true`,
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
    [baseUrl, namespace, workflow]
  );

  return memoizedTemplates;
};

const headerObjToCurlString = (header: Record<string, string>) =>
  Object.entries(header)
    .map(
      ([headerKey, headerValue]) => `\n -H '${headerKey}: ${headerValue}' \\`
    )
    .join("");

export const useCurlCommand = ({
  url,
  method,
  body,
}: {
  url: string;
  body: string;
  method: string;
}) => {
  const { isApiKeyRequired } = useApiKeyHandling();
  const apiKeyFromLocalstorage = useApiKey();

  const apiKeyHeader =
    isApiKeyRequired && apiKeyFromLocalstorage
      ? headerObjToCurlString(getAuthHeader(apiKeyFromLocalstorage))
      : "";

  const bodyEscaped = body.replace(/'/g, "\\'");

  return `curl '${url}' \\${apiKeyHeader}
  -X '${method}' \\
  --data-raw $'${bodyEscaped}'`;
};

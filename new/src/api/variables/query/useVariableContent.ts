import type { QueryFunctionContext } from "@tanstack/react-query";
import type { ResponseParser } from "../../apiFactory";
import { VarContentSchema } from "../schema";
import { apiFactory } from "../../apiFactory";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useQuery } from "@tanstack/react-query";
import { varKeys } from "..";

const variableResponseParser: ResponseParser = async ({ res, schema }) => {
  // Different from the default responseParser, for varibles we
  // don't want to try to parse the response as JSON because
  // it will always be treated as text to be displayed or edited
  // (even if it is JSON).
  const textResult = await res.text();
  if (!textResult) return schema.parse(null);

  // If response is not null, return it as 'body',
  // and also add the response headers (content type is needed later)
  // This is a workaround, in the new API we should return the
  // content type as part of the JSON response.
  const headers = Object.fromEntries(res.headers);
  return schema.parse({ body: textResult, headers });
};

const getVarContent = apiFactory({
  url: ({ namespace, name }: { namespace: string; name: string }) =>
    `/api/namespaces/${namespace}/vars/${name}`,
  method: "GET",
  schema: VarContentSchema,
  responseParser: variableResponseParser,
});

const fetchVarContent = async ({
  queryKey: [{ namespace, apiKey, name }],
}: QueryFunctionContext<ReturnType<(typeof varKeys)["varContent"]>>) =>
  getVarContent({
    apiKey,
    urlParams: { namespace, name },
  });

export const useVarContent = (name: string) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: varKeys.varContent(namespace, {
      apiKey: apiKey ?? undefined,
      name,
    }),
    queryFn: fetchVarContent,
    enabled: !!namespace,
  });
};

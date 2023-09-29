import { QueryFunctionContext } from "@tanstack/react-query";
import { TokenListSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { tokenKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

const getTokens = apiFactory({
  url: ({ namespace, baseUrl }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/tokens`,
  method: "GET",
  schema: TokenListSchema,
});

const fetchTokens = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof tokenKeys)["tokenList"]>>) =>
  getTokens({
    apiKey,
    urlParams: { namespace },
  });

export const useTokens = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: tokenKeys.tokenList(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchTokens,
    enabled: !!namespace,
  });
};

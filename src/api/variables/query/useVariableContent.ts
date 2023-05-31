import type { QueryFunctionContext } from "@tanstack/react-query";
import { VarContentSchema } from "../schema";
import { apiFactory } from "../../utils";
import { useApiKey } from "../../../util/store/apiKey";
import { useNamespace } from "../../../util/store/namespace";
import { useQuery } from "@tanstack/react-query";
import { varKeys } from "..";

const getVarContent = apiFactory({
  url: ({ namespace, name }: { namespace: string; name: string }) =>
    `/api/namespaces/${namespace}/vars/${name}`,
  method: "GET",
  schema: VarContentSchema,
});

const fetchVarContent = async ({
  queryKey: [{ namespace, apiKey, name }],
}: QueryFunctionContext<ReturnType<(typeof varKeys)["varContent"]>>) =>
  getVarContent({
    apiKey: apiKey,
    urlParams: { namespace, name },
    payload: undefined,
    headers: undefined,
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

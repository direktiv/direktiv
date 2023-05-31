import type { QueryFunctionContext } from "@tanstack/react-query";
import { VarListSchema } from "../schema";
import { apiFactory } from "../../utils";
import { useApiKey } from "../../../util/store/apiKey";
import { useNamespace } from "../../../util/store/namespace";
import { useQuery } from "@tanstack/react-query";
import { varKeys } from "..";

const getVars = apiFactory({
  // TODO: Pagination and Filter params
  url: ({ namespace }: { namespace: string }) =>
    `/api/namespaces/${namespace}/vars`,
  method: "GET",
  schema: VarListSchema,
});

const fetchVars = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof varKeys)["varList"]>>) =>
  getVars({
    apiKey: apiKey,
    urlParams: { namespace },
    payload: undefined,
    headers: undefined,
  });

export const useVars = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: varKeys.varList(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchVars,
    enabled: !!namespace,
  });
};

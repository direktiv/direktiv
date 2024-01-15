import type { QueryFunctionContext } from "@tanstack/react-query";
import { VarListSchema } from "../schema";
import { apiFactory } from "../../apiFactory";
import { useApiKey } from "../../../util/store/apiKey";
import { useNamespace } from "../../../util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";
import { varKeys } from "..";

const getVars = apiFactory({
  url: ({ namespace }: { namespace: string }) =>
    `/api/namespaces/${namespace}/vars`,
  method: "GET",
  schema: VarListSchema,
});

const fetchVars = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof varKeys)["varList"]>>) =>
  getVars({
    apiKey,
    urlParams: { namespace },
  });

export const useVars = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: varKeys.varList(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchVars,
    enabled: !!namespace,
  });
};

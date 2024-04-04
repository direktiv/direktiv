import { QueryFunctionContext } from "@tanstack/react-query";
import { VarContentSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";
import { varKeys } from "..";

export const getVariableDetails = apiFactory({
  url: ({ namespace, id }: { namespace: string; id: string }) =>
    `/api/v2/namespaces/${namespace}/variables/${id}`,
  method: "GET",
  schema: VarContentSchema,
});

export type VarDetailsType = Awaited<ReturnType<typeof getVariableDetails>>;

const fetchVarDetails = async ({
  queryKey: [{ apiKey, namespace, id }],
}: QueryFunctionContext<ReturnType<(typeof varKeys)["varDetails"]>>) =>
  getVariableDetails({
    apiKey,
    urlParams: { namespace, id },
  });

export const useVarDetails = (id: string) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: varKeys.varDetails(namespace, {
      apiKey: apiKey ?? undefined,
      id,
    }),
    queryFn: fetchVarDetails,
    enabled: !!namespace,
  });
};

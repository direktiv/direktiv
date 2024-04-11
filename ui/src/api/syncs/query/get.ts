import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { syncKeys } from "..";
import { syncListSchema } from "../schema";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

const getSyncs = apiFactory({
  url: ({ namespace }: { namespace: string }) =>
    `/api/v2/namespaces/${namespace}/syncs`,
  method: "GET",
  schema: syncListSchema,
});

const fetchSyncs = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof syncKeys)["syncsList"]>>) =>
  getSyncs({
    apiKey,
    urlParams: { namespace },
  });

export const useSyncs = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: syncKeys.syncsList(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchSyncs,
    enabled: !!namespace,
  });
};

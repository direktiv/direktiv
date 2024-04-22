import { SyncListSchema, SyncListSchemaType } from "../schema";

import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { syncKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

const getSyncs = apiFactory({
  url: ({ namespace }: { namespace: string }) =>
    `/api/v2/namespaces/${namespace}/syncs`,
  method: "GET",
  schema: SyncListSchema,
});

const fetchSyncs = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof syncKeys)["syncsList"]>>) =>
  getSyncs({
    apiKey,
    urlParams: { namespace },
  });

export const useSyncs = <T>({
  filter,
}: {
  filter: (apiResponse: SyncListSchemaType) => T;
}) => {
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
    select: (data) => filter(data),
  });
};

export const useListSyncs = () =>
  // copy array with spread notation to avoid mutating cached array
  useSyncs({
    filter: (apiResponse) => ({ data: [...apiResponse.data].reverse() }),
  });

export const useSyncDetail = (id: string) =>
  useSyncs({
    filter: (apiResponse) =>
      apiResponse.data.find((sybcObj) => sybcObj.id === id),
  });

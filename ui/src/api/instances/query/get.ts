import { FiltersObj, getFilterQuery } from "./utils";

import { InstancesListResponseSchema } from "../schema";
import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { buildSearchParamsString } from "~/api/utils";
import { instanceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

type InstanceListParams = {
  limit?: number;
  offset?: number;
  filters?: FiltersObj;
};

export const getInstances = apiFactory({
  url: ({
    namespace,
    baseUrl,
    filters,
    ...queryParams
  }: { baseUrl?: string; namespace: string } & InstanceListParams) => {
    let queryParamsString = buildSearchParamsString({ ...queryParams }, true);
    if (filters) {
      queryParamsString = queryParamsString.concat(getFilterQuery(filters));
    }
    queryParamsString = queryParamsString ? `?${queryParamsString}` : "";

    return `${
      baseUrl ?? ""
    }/api/v2/namespaces/${namespace}/instances${queryParamsString}`;
  },
  method: "GET",
  schema: InstancesListResponseSchema,
});

const fetchInstances = async ({
  queryKey: [{ apiKey, namespace, limit, offset, filters }],
}: QueryFunctionContext<ReturnType<(typeof instanceKeys)["instancesList"]>>) =>
  getInstances({
    apiKey,
    urlParams: { namespace, limit, offset, filters },
  });

export const useInstances = (params: InstanceListParams = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: instanceKeys.instancesList(namespace, {
      apiKey: apiKey ?? undefined,
      ...params,
    }),
    queryFn: fetchInstances,
    enabled: !!namespace,
  });
};

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

export const getInstanceList = apiFactory({
  url: ({
    namespace,
    baseUrl,
    filters,
    ...queryParams
  }: { baseUrl?: string; namespace: string } & InstanceListParams) => {
    let queryParamsString = buildSearchParamsString({
      ...queryParams,
    });

    if (filters) {
      queryParamsString = queryParamsString.concat(getFilterQuery(filters));
    }

    return `${
      baseUrl ?? ""
    }/api/v2/namespaces/${namespace}/instances/${queryParamsString}`;
  },
  method: "GET",
  schema: InstancesListResponseSchema,
});

const fetchInstanceList = async ({
  queryKey: [{ apiKey, namespace, limit, offset, filters }],
}: QueryFunctionContext<ReturnType<(typeof instanceKeys)["instancesList"]>>) =>
  getInstanceList({
    apiKey,
    urlParams: { namespace, limit, offset, filters },
  });

export const useInstanceList = (params: InstanceListParams = {}) => {
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
    queryFn: fetchInstanceList,
    enabled: !!namespace,
    select: (data) => data.data,
  });
};

import { InstancesListSchema } from "../schema";
import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { buildSearchParamsString } from "~/api/utils";
import { instanceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

type InstacenListParams = {
  limit?: number;
  offset?: number;
};

export const getInstanceList = apiFactory({
  url: ({
    namespace,
    baseUrl,
    ...queryParams
  }: { baseUrl?: string; namespace: string } & InstacenListParams) => {
    const queryParamsString = buildSearchParamsString({
      ...queryParams,
    });

    return `${
      baseUrl ?? ""
    }/api/v2/namespaces/${namespace}/instances/${queryParamsString}`;
  },
  method: "GET",
  schema: InstancesListSchema,
});

const fetchInstanceList = async ({
  queryKey: [{ apiKey, namespace, limit, offset }],
}: QueryFunctionContext<ReturnType<(typeof instanceKeys)["instancesList"]>>) =>
  getInstanceList({
    apiKey,
    urlParams: { namespace, limit, offset },
  });

export const useInstanceList = (params: InstacenListParams) => {
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

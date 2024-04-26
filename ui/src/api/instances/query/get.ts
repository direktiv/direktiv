import { InstancesListSchema } from "../schema";
import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { instanceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

export const getInstanceList = apiFactory({
  url: ({ namespace, baseUrl }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/instances/`,
  method: "GET",
  schema: InstancesListSchema,
});

const fetchInstanceList = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof instanceKeys)["instancesList"]>>) =>
  getInstanceList({
    apiKey,
    urlParams: { namespace },
  });

export const useInstanceList = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: instanceKeys.instancesList(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchInstanceList,
    enabled: !!namespace,
    select: (data) => data.data,
  });
};

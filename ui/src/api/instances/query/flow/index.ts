import { InstanceFlowResponseSchema } from "../../schema";
import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { instanceKeys } from "../..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

type InstanceDetailsQueryParams = {
  baseUrl?: string;
  namespace: string;
  instanceId: string;
};

const getInstanceFlow = apiFactory({
  url: ({ namespace, baseUrl, instanceId }: InstanceDetailsQueryParams) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/instances/${instanceId}/flow`,
  method: "GET",
  schema: InstanceFlowResponseSchema,
});

const fetchInstanceFlow = async ({
  queryKey: [{ apiKey, namespace, instanceId }],
}: QueryFunctionContext<ReturnType<(typeof instanceKeys)["instanceFlow"]>>) =>
  getInstanceFlow({
    apiKey,
    urlParams: { namespace, instanceId },
  });

export const useInstanceFlow = ({ instanceId }: { instanceId: string }) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: instanceKeys.instanceFlow(namespace, {
      apiKey: apiKey ?? undefined,
      instanceId,
    }),
    queryFn: fetchInstanceFlow,
    enabled: !!namespace,
    select: (data) => data,
  });
};

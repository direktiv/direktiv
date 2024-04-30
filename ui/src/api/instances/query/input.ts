import { InstancesInputResponseSchema } from "../schema";
import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { instanceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

export const getInstanceInput = apiFactory({
  url: ({
    namespace,
    baseUrl,
    instanceId,
  }: {
    baseUrl?: string;
    namespace: string;
    instanceId: string;
  }) =>
    `${
      baseUrl ?? ""
    }/api/v2/namespaces/${namespace}/instances/${instanceId}/input`,
  method: "GET",
  schema: InstancesInputResponseSchema,
});

const fetchInstanceInput = async ({
  queryKey: [{ apiKey, namespace, instanceId }],
}: QueryFunctionContext<ReturnType<(typeof instanceKeys)["instancesInput"]>>) =>
  getInstanceInput({
    apiKey,
    urlParams: { namespace, instanceId },
  });

export const useInstanceInput = ({ instanceId }: { instanceId: string }) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: instanceKeys.instancesInput(namespace, {
      apiKey: apiKey ?? undefined,
      instanceId,
    }),
    queryFn: fetchInstanceInput,
    enabled: !!namespace,
    select: (data) => data.data,
  });
};

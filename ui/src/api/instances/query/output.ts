import { InstancesOutputResponseSchema } from "../schema";
import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { instanceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

export const getInstanceOutput = apiFactory({
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
    }/api/v2/namespaces/${namespace}/instances/${instanceId}/output`,
  method: "GET",
  schema: InstancesOutputResponseSchema,
});

const fetchInstanceOutput = async ({
  queryKey: [{ apiKey, namespace, instanceId }],
}: QueryFunctionContext<ReturnType<(typeof instanceKeys)["instanceOutput"]>>) =>
  getInstanceOutput({
    apiKey,
    urlParams: { namespace, instanceId },
  });

export const useInstanceOutput = ({
  instanceId,
  enabled,
}: {
  instanceId: string;
  enabled?: boolean;
}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: instanceKeys.instanceOutput(namespace, {
      apiKey: apiKey ?? undefined,
      instanceId,
    }),
    queryFn: fetchInstanceOutput,
    enabled,
    select: (data) => data.data,
  });
};

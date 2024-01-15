import { QueryFunctionContext, useQuery } from "@tanstack/react-query";

import { InstancesOutputSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { instanceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";

export const getOutput = apiFactory({
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
    }/api/namespaces/${namespace}/instances/${instanceId}/output`,
  method: "GET",
  schema: InstancesOutputSchema,
});

const fetchOutput = async ({
  queryKey: [{ apiKey, namespace, instanceId }],
}: QueryFunctionContext<
  ReturnType<(typeof instanceKeys)["instancesOutput"]>
>) =>
  getOutput({
    apiKey,
    urlParams: { namespace, instanceId },
  });

export const useOutput = ({
  instanceId,
  enabled = true,
}: {
  instanceId: string;
  enabled?: boolean;
}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: instanceKeys.instancesOutput(namespace, {
      apiKey: apiKey ?? undefined,
      instanceId,
    }),
    queryFn: fetchOutput,
    enabled,
  });
};

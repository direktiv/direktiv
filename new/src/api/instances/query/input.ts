import { QueryFunctionContext, useQuery } from "@tanstack/react-query";

import { InstancesInputSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { instanceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";

export const getInput = apiFactory({
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
    }/api/namespaces/${namespace}/instances/${instanceId}/input`,
  method: "GET",
  schema: InstancesInputSchema,
});

const fetchInput = async ({
  queryKey: [{ apiKey, namespace, instanceId }],
}: QueryFunctionContext<ReturnType<(typeof instanceKeys)["instancesInput"]>>) =>
  getInput({
    apiKey,
    urlParams: { namespace, instanceId },
  });

export const useInput = ({ instanceId }: { instanceId: string }) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: instanceKeys.instancesInput(namespace, {
      apiKey: apiKey ?? undefined,
      instanceId,
    }),
    queryFn: fetchInput,
    enabled: !!namespace,
  });
};

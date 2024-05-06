import { InstanceDetailsResponseSchema } from "../../schema";
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

export const getInstanceDetails = apiFactory({
  url: ({ namespace, baseUrl, instanceId }: InstanceDetailsQueryParams) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/instances/${instanceId}`,
  method: "GET",
  schema: InstanceDetailsResponseSchema,
});

const fetchInstanceDetails = async ({
  queryKey: [{ apiKey, namespace, instanceId }],
}: QueryFunctionContext<
  ReturnType<(typeof instanceKeys)["instanceDetails"]>
>) =>
  getInstanceDetails({
    apiKey,
    urlParams: { namespace, instanceId },
  });

export const useInstanceDetails = ({ instanceId }: { instanceId: string }) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: instanceKeys.instanceDetails(namespace, {
      apiKey: apiKey ?? undefined,
      instanceId,
    }),
    queryFn: fetchInstanceDetails,
    enabled: !!namespace,
    select: (data) => data.data,
  });
};

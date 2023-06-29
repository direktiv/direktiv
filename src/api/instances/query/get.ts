import { QueryFunctionContext, useQuery } from "@tanstack/react-query";

import { InstancesListSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { instanceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";

export const getInstances = apiFactory({
  url: ({
    namespace,
    baseUrl,
    limit,
    offset,
  }: {
    baseUrl?: string;
    namespace: string;
    limit: number;
    offset: number;
  }) =>
    `${
      baseUrl ?? ""
    }/api/namespaces/${namespace}/instances?limit=${limit}&offset=${offset}`,
  method: "GET",
  schema: InstancesListSchema,
});

const fetchInstances = async ({
  queryKey: [{ apiKey, namespace, limit, offset }],
}: QueryFunctionContext<ReturnType<(typeof instanceKeys)["instancesList"]>>) =>
  getInstances({
    apiKey,
    urlParams: { namespace, limit, offset },
  });

export const useInstances = ({
  limit,
  offset,
}: {
  limit: number;
  offset: number;
}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: instanceKeys.instancesList(namespace, {
      apiKey: apiKey ?? undefined,
      limit,
      offset,
    }),
    queryFn: fetchInstances,
    enabled: !!namespace,
  });
};

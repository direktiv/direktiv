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
    filter,
  }: {
    baseUrl?: string;
    namespace: string;
    limit: number;
    offset: number;
    filter?: string;
  }) =>
    `${
      baseUrl ?? ""
    }/api/namespaces/${namespace}/instances?limit=${limit}&offset=${offset}${
      filter ?? filter
    }`,
  method: "GET",
  schema: InstancesListSchema,
});

const fetchInstances = async ({
  queryKey: [{ apiKey, namespace, limit, offset, filter }],
}: QueryFunctionContext<ReturnType<(typeof instanceKeys)["instancesList"]>>) =>
  getInstances({
    apiKey,
    urlParams: { namespace, limit, offset, filter },
  });

export const useInstances = ({
  limit,
  offset,
  filter,
}: {
  limit: number;
  offset: number;
  filter: string;
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
      filter,
    }),
    queryFn: fetchInstances,
    enabled: !!namespace,
  });
};

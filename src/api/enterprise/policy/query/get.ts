import { QueryFunctionContext, useQuery } from "@tanstack/react-query";

import { PolicySchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { policyKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";

const getPolicy = apiFactory({
  url: ({ namespace, baseUrl }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/v2/namespace/${namespace}/policy`,
  method: "GET",
  schema: PolicySchema,
});

const fetchPolicy = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof policyKeys)["get"]>>) =>
  getPolicy({
    apiKey,
    urlParams: {
      namespace,
    },
  });

export const usePolicy = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: policyKeys.get(namespace, { apiKey: apiKey ?? undefined }),
    queryFn: fetchPolicy,
  });
};

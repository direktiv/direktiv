import { PolicySchema } from "../schema";
import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { policyKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

const getPolicy = apiFactory({
  url: ({ namespace, baseUrl }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/policy`,
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

  return useQueryWithPermissions({
    queryKey: policyKeys.get(namespace, { apiKey: apiKey ?? undefined }),
    queryFn: fetchPolicy,
  });
};

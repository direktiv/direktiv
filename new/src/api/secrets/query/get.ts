import type { QueryFunctionContext } from "@tanstack/react-query";
import { SecretListSchema } from "../schema";
import { apiFactory } from "../../apiFactory";
import { secretKeys } from "..";
import { useApiKey } from "../../../util/store/apiKey";
import { useNamespace } from "../../../util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

const getSecrets = apiFactory({
  url: ({ namespace }: { namespace: string }) =>
    `/api/namespaces/${namespace}/secrets`,
  method: "GET",
  schema: SecretListSchema,
});

const fetchSecrets = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof secretKeys)["secretsList"]>>) =>
  getSecrets({
    apiKey,
    urlParams: { namespace },
  });

export const useSecrets = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: secretKeys.secretsList(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchSecrets,
    enabled: !!namespace,
  });
};

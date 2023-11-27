import { LintSchema } from "../schema";
import type { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "../../apiFactory";
import { lintingKeys } from "..";
import { useApiKey } from "../../../util/store/apiKey";
import { useNamespace } from "../../../util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

const getNamespaceLinting = apiFactory({
  url: ({ namespace }: { namespace: string }) =>
    `/api/namespaces/${namespace}/lint`,
  method: "GET",
  schema: LintSchema,
});

const fetchLinting = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof lintingKeys)["getLinting"]>>) =>
  getNamespaceLinting({
    apiKey,
    urlParams: { namespace },
  });

export const useNamespaceLinting = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: lintingKeys.getLinting(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchLinting,
    enabled: !!namespace,
  });
};

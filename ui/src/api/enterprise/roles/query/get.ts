import { QueryFunctionContext } from "@tanstack/react-query";
import { RolesListSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { roleKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

const getRoles = apiFactory({
  url: ({ namespace, baseUrl }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/roles`,
  method: "GET",
  schema: RolesListSchema,
});

const fetchRoles = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof roleKeys)["roleList"]>>) =>
  getRoles({
    apiKey,
    urlParams: { namespace },
  });

export const useRoles = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: roleKeys.roleList(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchRoles,
    enabled: !!namespace,
  });
};

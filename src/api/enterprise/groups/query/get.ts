import { QueryFunctionContext, useQuery } from "@tanstack/react-query";

import { GroupsListSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { groupKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";

export const getGroups = apiFactory({
  url: ({ namespace, baseUrl }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/groups`,
  method: "GET",
  schema: GroupsListSchema,
});

const fetchGroups = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof groupKeys)["groupList"]>>) =>
  getGroups({
    apiKey,
    urlParams: { namespace },
  });

export const useGroups = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: groupKeys.groupList(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchGroups,
    enabled: !!namespace,
  });
};

import { QueryFunctionContext, useQuery } from "@tanstack/react-query";

import { GroupslistSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { groupsKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";

export const getGroups = apiFactory({
  url: ({ namespace, baseUrl }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/groups`,
  method: "GET",
  schema: GroupslistSchema,
});

const fetchGroups = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof groupsKeys)["groupList"]>>) =>
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
    queryKey: groupsKeys.groupList(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchGroups,
    enabled: !!namespace,
  });
};

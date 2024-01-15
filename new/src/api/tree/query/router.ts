import { forceLeadingSlash, sortByRef } from "../utils";

import type { QueryFunctionContext } from "@tanstack/react-query";
import { RouterSchema } from "../schema/node";
import { apiFactory } from "../../apiFactory";
import { treeKeys } from "../";
import { useApiKey } from "../../../util/store/apiKey";
import { useNamespace } from "../../../util/store/namespace";
import { useQuery } from "@tanstack/react-query";

const getRouter = apiFactory({
  url: ({ namespace, path }: { namespace: string; path?: string }) =>
    `/api/namespaces/${namespace}/tree${forceLeadingSlash(path)}?op=router`,
  method: "GET",
  schema: RouterSchema,
});

const fetchRouter = async ({
  queryKey: [{ apiKey, namespace, path }],
}: QueryFunctionContext<ReturnType<(typeof treeKeys)["router"]>>) =>
  getRouter({
    apiKey,
    urlParams: {
      namespace,
      path,
    },
  });

export const useRouter = ({
  path,
}: {
  path?: string;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: treeKeys.router(namespace, {
      apiKey: apiKey ?? undefined,
      path,
    }),
    queryFn: fetchRouter,
    enabled: !!namespace,
    // TODO: waiting for DIR-576 to get fixed
    select: (data) => ({
      ...data,
      routes: [...data.routes.sort(sortByRef)],
    }),
  });
};

import { QueryFunctionContext, useQuery } from "@tanstack/react-query";

import { MirrorInfoSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { treeKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";

const getMirrorInfo = apiFactory({
  url: ({ namespace }: { namespace: string; path?: string }) =>
    `/api/namespaces/${namespace}/tree?op=mirror-info`,
  method: "GET",
  schema: MirrorInfoSchema,
});

const fetchMirrorInfo = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof treeKeys)["mirrorInfo"]>>) =>
  getMirrorInfo({
    apiKey,
    urlParams: { namespace },
  });

export const useMirrorInfo = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: treeKeys.mirrorInfo(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchMirrorInfo,
    enabled: !!namespace,
  });
};

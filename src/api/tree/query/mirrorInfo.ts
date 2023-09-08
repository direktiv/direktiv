import { MirrorInfoSchema } from "../schema/mirror";
import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { treeKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

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

  return useQueryWithPermissions({
    queryKey: treeKeys.mirrorInfo(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchMirrorInfo,
    enabled: !!namespace,
  });
};

export const useMirrorActivity = ({ id }: { id: string }) => {
  const {
    data: mirrorInfo,
    isAllowed,
    noPermissionMessage,
    isFetched,
  } = useMirrorInfo();

  const data = mirrorInfo?.activities.results.find((item) => item.id === id);
  return { data, isAllowed, noPermissionMessage, isFetched };
};

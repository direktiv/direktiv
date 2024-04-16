import { MirrorInfoSchema } from "../../namespaces/schema/mirror";
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
  }).then((res) => ({
    ...res,
    activities: {
      ...res.activities,
      // This should be changed in the backend in [DIR-833]
      // reverse the order of activities (newer first)
      results: [...(res.activities.results ?? []).reverse()],
    },
  }));

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

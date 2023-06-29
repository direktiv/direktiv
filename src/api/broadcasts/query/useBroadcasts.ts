import { BroadcastsResponseSchema } from "../schema";
import type { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { broadcastKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useQuery } from "@tanstack/react-query";

const getBroadcasts = apiFactory({
  url: ({ namespace }: { namespace: string }) =>
    `/api/namespaces/${namespace}/config`,
  method: "GET",
  schema: BroadcastsResponseSchema,
});

const fetchBroadcasts = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof broadcastKeys)["broadcasts"]>>) =>
  getBroadcasts({
    apiKey,
    urlParams: { namespace },
  });

export const useBroadcasts = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: broadcastKeys.broadcasts(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchBroadcasts,
    enabled: !!namespace,
  });
};

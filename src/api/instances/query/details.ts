import { InstancesDetailSchema, InstancesDetailSchemaType } from "../schema";
import {
  QueryFunctionContext,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";

import { apiFactory } from "~/api/apiFactory";
import { instanceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useStreaming } from "~/api/streaming";

export const getInstanceDetails = apiFactory({
  url: ({
    namespace,
    baseUrl,
    instanceId,
  }: {
    baseUrl?: string;
    namespace: string;
    instanceId: string;
  }) => `${baseUrl ?? ""}/api/namespaces/${namespace}/instances/${instanceId}`,
  method: "GET",
  schema: InstancesDetailSchema,
});

const fetchInstanceDetails = async ({
  queryKey: [{ apiKey, namespace, instanceId }],
}: QueryFunctionContext<ReturnType<(typeof instanceKeys)["instanceDetail"]>>) =>
  getInstanceDetails({
    apiKey,
    urlParams: { namespace, instanceId },
  });

export const useInstanceDetails = (
  { instanceId }: { instanceId: string },
  { streaming }: { streaming?: boolean } = {}
) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  const streamingUrl = `/api/namespaces/${namespace}/instances/${instanceId}`;
  useStreaming({
    url: streamingUrl,
    enabled: streaming,
    onMessage: (msg) => {
      if (!msg.data) return null;

      let msgJson = null;
      try {
        // try to parse the response as json
        msgJson = JSON.parse(msg.data);
      } catch (e) {
        console.error(
          `error parsing streaming result from ${streamingUrl} as json`,
          msg.data
        );
        return;
      }

      const parsedResult = InstancesDetailSchema.safeParse(msgJson);

      if (parsedResult.success === false) {
        console.error(`error parsing streaming result for ${streamingUrl}`);
        return;
      }

      queryClient.setQueryData<InstancesDetailSchemaType>(
        instanceKeys.instanceDetail(namespace, {
          apiKey: apiKey ?? undefined,
          instanceId,
        }),
        () => parsedResult.data
      );
    },
  });

  return useQuery({
    queryKey: instanceKeys.instanceDetail(namespace, {
      apiKey: apiKey ?? undefined,
      instanceId,
    }),
    queryFn: fetchInstanceDetails,
    // disable queryFn when streaming is enabled (to avoid duplicate requests)
    enabled: !streaming,
  });
};

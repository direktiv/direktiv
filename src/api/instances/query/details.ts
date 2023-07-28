import { QueryFunctionContext, useQuery } from "@tanstack/react-query";
import { useEffect, useRef } from "react";

import { InstancesDetailSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { instanceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";

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

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  useStreaming({
    url: `/api/namespaces/${namespace}/instances/${instanceId}`,
    enabled: streaming,
    onMessage: (msg) => {
      console.warn("🚀 received a message", msg);
    },
  });

  return useQuery({
    queryKey: instanceKeys.instanceDetail(namespace, {
      apiKey: apiKey ?? undefined,
      instanceId,
    }),
    queryFn: fetchInstanceDetails,
    enabled: !!namespace,
  });
};

export const useStreaming = ({
  url,
  onOpen,
  onMessage,
  onError,
  enabled,
}: {
  url: string;
  onOpen?: (e: Event) => void;
  onMessage?: (e: MessageEvent) => void;
  onError?: (e: Event) => void;
  enabled?: boolean;
}) => {
  const eventSource = useRef<EventSource | null>(null);

  const stopStreaming = () => {
    eventSource.current?.close();
    eventSource.current = null;
  };

  const startSteaming = () => {
    if (enabled && eventSource.current === null) {
      const listener = new EventSource(url);
      eventSource.current = listener;
      if (onOpen) listener.onopen = onOpen;
      if (onError) listener.onerror = onError;
      if (onMessage) listener.onmessage = onMessage;
    }
  };

  useEffect(() => {
    startSteaming();
    return () => {
      stopStreaming();
    };
  });
};

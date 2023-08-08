import { NamespaceLogListSchema, NamespaceLogListSchemaType } from "../schema";
import {
  QueryFunctionContext,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";

import { apiFactory } from "~/api/apiFactory";
import { namespaceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useStreaming } from "~/api/streaming";

export const getInstanceDetails = apiFactory({
  url: ({ namespace, baseUrl }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/namespaces/${namespace}/logs`,
  method: "GET",
  schema: NamespaceLogListSchema,
});

const fetchInstanceDetails = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof namespaceKeys)["logs"]>>) =>
  getInstanceDetails({
    apiKey,
    urlParams: { namespace },
  });

export const useNamespaceLogsStream = ({
  enabled = true,
}: { enabled?: boolean } = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useStreaming({
    url: `/api/namespaces/${namespace}/logs`,
    apiKey: apiKey ?? undefined,
    enabled,
    schema: NamespaceLogListSchema,
    onMessage: (msg) => {
      queryClient.setQueryData<NamespaceLogListSchemaType>(
        namespaceKeys.logs(namespace, { apiKey: apiKey ?? undefined }),
        () => msg
      );
    },
  });
};

export const useNamespacelogs = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: namespaceKeys.logs(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchInstanceDetails,
    enabled: !!namespace,
  });
};

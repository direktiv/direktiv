import {
  BroadcastsPatchSchemaType,
  BroadcastsResponseSchema,
  BroadcastsResponseSchemaType,
} from "../schema";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { apiFactory } from "~/api/apiFactory";
import { broadcastKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";

const updateBroadcasts = apiFactory({
  url: ({ namespace }: { namespace: string }) =>
    `/api/namespaces/${namespace}/config`,
  method: "PATCH",
  schema: BroadcastsResponseSchema,
});

export const useUpdateBroadcasts = ({
  onSuccess,
}: {
  onSuccess?: (data: BroadcastsResponseSchemaType) => void;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  const mutationFn = ({
    payload,
  }: {
    payload: BroadcastsPatchSchemaType; // Partial<BroadcastsSchemaType>?
  }) =>
    updateBroadcasts({
      apiKey: apiKey ?? undefined,
      payload,
      urlParams: {
        namespace,
      },
    });

  return useMutation({
    mutationFn,
    onSuccess: (data) => {
      queryClient.setQueryData<BroadcastsResponseSchemaType>(
        broadcastKeys.broadcasts(namespace, { apiKey: apiKey ?? undefined }),
        data
      );
      onSuccess?.(data);
    },
  });
};

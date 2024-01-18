import {
  BroadcastsPatchSchemaType,
  BroadcastsResponseSchema,
  BroadcastsResponseSchemaType,
} from "../schema";

import { apiFactory } from "~/api/apiFactory";
import { broadcastKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";

export const updateBroadcasts = apiFactory({
  url: ({ baseUrl, namespace }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/namespaces/${namespace}/config`,
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

  const mutationFn = ({ payload }: { payload: BroadcastsPatchSchemaType }) =>
    updateBroadcasts({
      apiKey: apiKey ?? undefined,
      payload,
      urlParams: {
        namespace,
      },
    });

  return useMutationWithPermissions({
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

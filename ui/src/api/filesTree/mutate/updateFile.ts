import { NodeSchemaType, UpdateFileSchemaType } from "../schema";

import { forceLeadingSlash } from "~/api/tree/utils";
import { getMessageFromApiError } from "~/api/errorHandling";
import { patchNode } from "./patchNode";
import { pathKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";

export const useUpdateFile = ({
  onSuccess,
  onError,
}: {
  onSuccess?: () => void;
  onError?: (e: string | undefined) => void;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutationWithPermissions({
    mutationFn: ({
      node,
      file,
    }: {
      node: NodeSchemaType;
      file: UpdateFileSchemaType;
    }) =>
      patchNode({
        apiKey: apiKey ?? undefined,
        payload: file,
        urlParams: {
          path: node.path,
          namespace,
        },
      }),
    onSuccess(data) {
      queryClient.invalidateQueries(
        pathKeys.paths(namespace, {
          apiKey: apiKey ?? undefined,
          path: forceLeadingSlash(data.data.path),
        })
      );
      onSuccess?.();
    },
    onError: (e) => {
      onError?.(getMessageFromApiError(e));
    },
  });
};

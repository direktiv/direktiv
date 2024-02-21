import { UpdateFileSchemaType } from "../schema";
import { fileKeys } from "..";
import { forceLeadingSlash } from "~/api/tree/utils";
import { getMessageFromApiError } from "~/api/errorHandling";
import { patchFile } from "./patchFile";
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
      path,
      payload,
    }: {
      path: string;
      payload: UpdateFileSchemaType;
    }) =>
      patchFile({
        apiKey: apiKey ?? undefined,
        payload,
        urlParams: {
          path,
          namespace,
        },
      }),
    onSuccess(data) {
      queryClient.invalidateQueries(
        fileKeys.children(namespace, {
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

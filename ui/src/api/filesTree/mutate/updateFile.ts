import {
  NodeSchemaType,
  UpdateFileSchemaType,
  getFilenameFromPath,
  getParentFromPath,
} from "../schema";

import { getMessageFromApiError } from "~/api/errorHandling";
import { patchNode } from "./patchNode";
import { pathKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

export const useUpdateFile = ({
  onSuccess,
  onError,
}: {
  onSuccess?: () => void;
  onError?: (e: string | undefined) => void;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const { t } = useTranslation();

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
    onSuccess(data, variables) {
      queryClient.invalidateQueries(
        pathKeys.paths(namespace, {
          apiKey: apiKey ?? undefined,
          path: getParentFromPath(data.data.path),
        })
      );
      toast({
        title: t("api.tree.mutate.renameNode.success.title", {
          type: variables.node.type === "workflow" ? "workflow" : "directory",
        }),
        description: t("api.tree.mutate.renameNode.success.description", {
          name: getFilenameFromPath(variables.node.path),
        }),
        variant: "success",
      });
      onSuccess?.();
    },
    onError: (e) => {
      onError?.(getMessageFromApiError(e));
    },
  });
};

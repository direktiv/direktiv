import {
  NodeSchemaType,
  PathDeletedSchema,
  getFilenameFromPath,
  getParentFromPath,
} from "../schema";

import { apiFactory } from "~/api/apiFactory";
import { forceLeadingSlash } from "~/api/tree/utils";
import { pathKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const deleteNode = apiFactory({
  url: ({ namespace, path }: { namespace: string; path: string }) =>
    `/api/v2/namespaces/${namespace}/files-tree${forceLeadingSlash(path)}`,
  method: "DELETE",
  schema: PathDeletedSchema,
});

export const useDeletePath = ({
  onSuccess,
}: { onSuccess?: () => void } = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutationWithPermissions({
    mutationFn: ({ node }: { node: NodeSchemaType }) =>
      deleteNode({
        apiKey: apiKey ?? undefined,
        urlParams: {
          path: node.path,
          namespace,
        },
      }),
    onSuccess(_, variables) {
      queryClient.invalidateQueries(
        pathKeys.paths(namespace, {
          apiKey: apiKey ?? undefined,
          path: getParentFromPath(variables.node.path),
        })
      );
      toast({
        title: t("api.tree.mutate.deleteNode.success.title", {
          type: variables.node.type === "workflow" ? "workflow" : "directory",
        }),
        description: t("api.tree.mutate.deleteNode.success.description", {
          name: getFilenameFromPath(variables.node.path),
        }),
        variant: "success",
      });
      onSuccess?.();
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.tree.mutate.deleteNode.error.description"),
        variant: "error",
      });
    },
  });
};

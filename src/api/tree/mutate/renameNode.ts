import { NodeRenameSchema, NodeSchemaType } from "../schema/node";
import { forceLeadingSlash, removeLeadingSlash } from "../utils";

import { apiFactory } from "~/api/apiFactory";
import { treeKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const renameNode = apiFactory({
  url: ({ namespace, path }: { namespace: string; path: string }) =>
    `/api/namespaces/${namespace}/tree${forceLeadingSlash(
      path
    )}/?op=rename-node`,
  method: "POST",
  schema: NodeRenameSchema,
});

export const useRenameNode = ({
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
    mutationFn: ({
      node,
      newName,
    }: {
      node: NodeSchemaType;
      newName: string;
    }) =>
      renameNode({
        apiKey: apiKey ?? undefined,
        payload: {
          new: `${removeLeadingSlash(node.parent)}/${newName}`,
        },
        urlParams: {
          path: node.path,
          namespace,
        },
      }),
    onSuccess(data, variables) {
      queryClient.invalidateQueries(
        treeKeys.nodeContent(namespace, {
          apiKey: apiKey ?? undefined,
          path: data.node.parent,
        })
      );

      toast({
        title: t("api.tree.mutate.renameNode.success.title", {
          type: variables.node.type === "workflow" ? "workflow" : "directory",
        }),
        description: t("api.tree.mutate.renameNode.success.description", {
          name: variables.node.name,
        }),
        variant: "success",
      });
      onSuccess?.();
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.tree.mutate.renameNode.error.description"),
        variant: "error",
      });
    },
  });
};

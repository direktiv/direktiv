import {
  NodeDeletedSchema,
  NodeListSchemaType,
  NodeSchemaType,
} from "../schema/node";

import { apiFactory } from "~/api/apiFactory";
import { forceLeadingSlash } from "../utils";
import { treeKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const updateCache = (
  oldData: NodeListSchemaType | undefined,
  variables: Parameters<ReturnType<typeof useDeleteNode>["mutate"]>[0]
) => {
  if (!oldData) return undefined;
  const oldChildren = oldData?.children;
  return {
    ...oldData,
    ...(oldChildren
      ? {
          children: {
            ...oldChildren,
            results: oldChildren?.results.filter(
              (child) => child.name !== variables.node.name
            ),
          },
        }
      : {}),
  };
};

const deleteNode = apiFactory({
  url: ({ namespace, path }: { namespace: string; path: string }) =>
    `/api/namespaces/${namespace}/tree${forceLeadingSlash(
      path
    )}/?op=delete-node&recursive=true`,
  method: "DELETE",
  schema: NodeDeletedSchema,
});

export const useDeleteNode = ({
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
      queryClient.setQueryData<NodeListSchemaType>(
        treeKeys.nodeContent(namespace, {
          apiKey: apiKey ?? undefined,
          path: variables.node.parent,
        }),
        (oldData) => updateCache(oldData, variables)
      );
      toast({
        title: t("api.tree.mutate.deleteNode.success.title", {
          type: variables.node.type === "workflow" ? "workflow" : "directory",
        }),
        description: t("api.tree.mutate.deleteNode.success.description", {
          name: variables.node.name,
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

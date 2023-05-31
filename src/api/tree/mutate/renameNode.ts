import {
  NodeListSchemaType,
  NodeRenameSchema,
  NodeSchemaType,
} from "../schema";
import {
  forceLeadingSlash,
  removeLeadingSlash,
  removeTrailingSlash,
} from "../utils";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { apiFactory } from "~/api/utils";
import { treeKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
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
    throw new Error(t("api.generic.undefinedNamespace"));
  }

  return useMutation({
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
          namespace: namespace,
        },
      }),
    onSuccess(data, variables) {
      queryClient.setQueryData<NodeListSchemaType>(
        treeKeys.nodeContent(namespace, {
          apiKey: apiKey ?? undefined,
          path: variables.node.parent,
        }),
        (oldData) => {
          if (!oldData) return undefined;
          const oldChildren = oldData?.children;
          return {
            ...oldData,
            ...(oldChildren
              ? {
                  children: {
                    ...oldChildren,
                    results: oldChildren?.results.map((child) => {
                      if (child.path === variables.node.path) {
                        return {
                          ...data.node,
                          // there is a bug in the API where the returned data after
                          // a rename is wrong. The name and updatedAt are not updated
                          // and the parent will have a trailing slash, which it does
                          // not have in the original data from the tree list
                          name: variables.newName,
                          parent:
                            variables.node.parent === "/"
                              ? "/"
                              : removeTrailingSlash(variables.node.parent),
                          updatedAt: new Date().toISOString(),
                        };
                      }
                      return child;
                    }),
                  },
                }
              : {}),
          };
        }
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

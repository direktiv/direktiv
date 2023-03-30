import {
  NodeSchemaType,
  TreeListSchemaType,
  TreeNodeRenameSchema,
} from "../schema";
import { apiFactory, defaultKeys } from "../../utils";
import { forceSlashIfPath, removeSlashIfPath } from "../utils";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { treeKeys } from "..";
import { useApiKey } from "../../../util/store/apiKey";
import { useNamespace } from "../../../util/store/namespace";
import { useToast } from "../../../design/Toast";

const renameNode = apiFactory({
  pathFn: ({ namespace, path }: { namespace: string; path: string }) =>
    `/api/namespaces/${namespace}/tree${forceSlashIfPath(
      path
    )}/?op=rename-node`,
  method: "POST",
  schema: TreeNodeRenameSchema,
});

export const useRenameNode = ({
  onSuccess,
}: { onSuccess?: () => void } = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
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
        params: {
          new: `${removeSlashIfPath(node.parent)}/${newName}`,
        },
        pathParams: {
          path: node.path,
          namespace: namespace,
        },
      }),
    onSuccess(data, variables) {
      queryClient.setQueryData<TreeListSchemaType>(
        treeKeys.all(
          apiKey ?? defaultKeys.apiKey,
          namespace,
          variables.node.parent ?? ""
        ),
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
                          // there is a bug in the API where the returned data after a rename
                          // don't update the name and updatedAt
                          name: variables.newName,
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
        title: `${
          variables.node.type === "workflow" ? "workflow" : "directory"
        } renamed`,
        description: `${variables.node.name} was renamed`,
        variant: "success",
      });
      onSuccess?.();
    },
    onError: () => {
      toast({
        title: "An error occurred",
        description: "could not rename 😢",
        variant: "error",
      });
    },
  });
};

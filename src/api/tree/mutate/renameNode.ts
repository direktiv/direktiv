import {
  NodeSchemaType,
  TreeListSchemaType,
  TreeNodeRenameSchema,
} from "../schema";
import { apiFactory, defaultKeys } from "../../utils";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { forceSlashIfPath } from "../utils";
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
      newPath,
    }: {
      node: NodeSchemaType;
      newPath: string;
    }) =>
      renameNode({
        apiKey: apiKey ?? undefined,
        params: {
          new: newPath,
        },
        pathParams: {
          path: node.path,
          namespace: namespace,
        },
      }),
    onSuccess(_, variables) {
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
                    results: oldChildren?.results.filter(
                      (child) => child.name !== variables.node.name
                    ),
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

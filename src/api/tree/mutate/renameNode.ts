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

import { apiFactory } from "../../utils";
import { treeKeys } from "..";
import { useApiKey } from "../../../util/store/apiKey";
import { useNamespace } from "../../../util/store/namespace";
import { useToast } from "../../../design/Toast";

const renameNode = apiFactory({
  pathFn: ({ namespace, path }: { namespace: string; path: string }) =>
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
          new: `${removeLeadingSlash(node.parent)}/${newName}`,
        },
        pathParams: {
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

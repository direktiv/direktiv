import {
  NodeDeletedSchema,
  NodeListSchemaType,
  NodeSchemaType,
} from "../schema";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { apiFactory } from "../../utils";
import { forceLeadingSlash } from "../utils";
import { treeKeys } from "..";
import { useApiKey } from "../../../util/store/apiKey";
import { useNamespace } from "../../../util/store/namespace";
import { useToast } from "../../../design/Toast";

const deleteNode = apiFactory({
  pathFn: ({ namespace, path }: { namespace: string; path: string }) =>
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

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutation({
    mutationFn: ({ node }: { node: NodeSchemaType }) =>
      deleteNode({
        apiKey: apiKey ?? undefined,
        params: undefined,
        pathParams: {
          path: node.path,
          namespace: namespace,
        },
      }),
    onSuccess(_, variables) {
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
        } deleted`,
        description: `${variables.node.name} was deleted`,
        variant: "success",
      });
      onSuccess?.();
    },
    onError: () => {
      toast({
        title: "An error occurred",
        description: "could not delete 😢",
        variant: "error",
      });
    },
  });
};

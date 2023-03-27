import { NodeSchemaType, TreeNodeDeletedSchema } from "../schema";

import { apiFactory } from "../../utils";
import { forceSlashIfPath } from "../utils";
import { useApiKey } from "../../../util/store/apiKey";
import { useMutation } from "@tanstack/react-query";
import { useNamespace } from "../../../util/store/namespace";
import { useToast } from "../../../componentsNext/Toast";

const deleteNode = apiFactory({
  pathFn: ({ namespace, path }: { namespace: string; path: string }) =>
    `/api/namespaces/${namespace}/tree${forceSlashIfPath(
      path
    )}/?op=delete-node&recursive=true`,
  method: "DELETE",
  schema: TreeNodeDeletedSchema,
});

export const useDeleteNode = ({
  onSuccess,
}: { onSuccess?: () => void } = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();

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
      toast({
        title: `${
          variables.node.type === "workflow" ? "workflow" : "directory"
        } deleted`,
        description: `${variables.node.parent} was deleted`,
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

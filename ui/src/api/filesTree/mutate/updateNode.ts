import {
  NodePatchedSchema,
  NodeSchemaType,
  PatchNodeSchemaType,
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

const updateNode = apiFactory({
  url: ({ namespace, path }: { namespace: string; path: string }) =>
    `/api/v2/namespaces/${namespace}/files-tree${forceLeadingSlash(path)}`,
  method: "PATCH",
  schema: NodePatchedSchema,
});

export const useUpdateNode = ({
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
      file,
    }: {
      node: NodeSchemaType;
      file: PatchNodeSchemaType;
    }) =>
      updateNode({
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
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.tree.mutate.renameNode.error.description"),
        variant: "error",
      });
    },
  });
};

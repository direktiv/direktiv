import {
  NodeDeletedSchema,
  RevisionsListSchemaType,
  TagsListSchemaType,
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

const deleteTag = apiFactory({
  url: ({
    namespace,
    path,
    tag,
  }: {
    namespace: string;
    path: string;
    tag: string;
  }) =>
    `/api/namespaces/${namespace}/tree${forceLeadingSlash(
      path
    )}/?op=untag&ref=${tag}`,
  method: "POST",
  schema: NodeDeletedSchema,
});

export const useDeleteTag = ({
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
    mutationFn: ({ path, tag }: { path: string; tag: string }) =>
      deleteTag({
        apiKey: apiKey ?? undefined,
        urlParams: {
          path,
          namespace,
          tag,
        },
      }),
    onSuccess(_, variables) {
      toast({
        title: t("api.tree.mutate.deleteTag.success.title"),
        description: t("api.tree.mutate.deleteTag.success.description", {
          name: variables.tag,
        }),
        variant: "success",
      });
      // update revisions list cache
      queryClient.setQueryData<RevisionsListSchemaType>(
        treeKeys.revisionsList(namespace, {
          apiKey: apiKey ?? undefined,
          path: variables.path,
        }),
        (oldData) => {
          if (!oldData) return undefined;
          const oldRevisions = oldData?.results;
          return {
            ...oldData,
            ...(oldRevisions
              ? {
                  results: oldRevisions?.filter(
                    (child) => child.name !== variables.tag
                  ),
                }
              : {}),
          };
        }
      );
      // update tags list cache
      queryClient.setQueryData<TagsListSchemaType>(
        treeKeys.tagsList(namespace, {
          apiKey: apiKey ?? undefined,
          path: variables.path,
        }),
        (oldData) => {
          if (!oldData) return undefined;
          const oldTags = oldData?.results;
          return {
            ...oldData,
            ...(oldTags
              ? {
                  results: oldTags?.filter(
                    (child) => child.name !== variables.tag
                  ),
                }
              : {}),
          };
        }
      );
      onSuccess?.();
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.tree.mutate.deleteTag.error.description"),
        variant: "error",
      });
    },
  });
};

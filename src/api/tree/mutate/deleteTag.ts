import {
  NodeDeletedSchema,
  RevisionsListSchemaType,
  TagsListSchemaType,
} from "../schema";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { apiFactory } from "~/api/utils";
import { forceLeadingSlash } from "../utils";
import { treeKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useToast } from "~/design/Toast";

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

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutation({
    mutationFn: ({ path, tag }: { path: string; tag: string }) =>
      deleteTag({
        apiKey: apiKey ?? undefined,
        payload: undefined,
        headers: undefined,
        urlParams: {
          path,
          namespace: namespace,
          tag,
        },
      }),
    onSuccess(_, variables) {
      toast({
        title: `tag deleted`,
        description: `tag ${variables.tag} was deleted`,
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
        title: "An error occurred",
        description: "could not delete 😢",
        variant: "error",
      });
    },
  });
};

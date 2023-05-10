import { RevisionsListSchemaType, TreeNodeDeletedSchema } from "../schema";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { apiFactory } from "../../utils";
import { forceLeadingSlash } from "../utils";
import { treeKeys } from "..";
import { useApiKey } from "../../../util/store/apiKey";
import { useNamespace } from "../../../util/store/namespace";
import { useToast } from "../../../design/Toast";

const deleteRevision = apiFactory({
  pathFn: ({
    namespace,
    path,
    revision,
  }: {
    namespace: string;
    path: string;
    revision: string;
  }) =>
    `/api/namespaces/${namespace}/tree${forceLeadingSlash(
      path
    )}/?op=delete-revision&ref=${revision}`,
  method: "POST",
  schema: TreeNodeDeletedSchema,
});

export const useDeleteRevision = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutation({
    mutationFn: ({ path, revision }: { path: string; revision: string }) =>
      deleteRevision({
        apiKey: apiKey ?? undefined,
        params: undefined,
        pathParams: {
          path,
          namespace: namespace,
          revision,
        },
      }),
    onSuccess(_, variables) {
      toast({
        title: `revision deleted`,
        description: `revision ${variables.revision.slice(0, 8)} was deleted`,
        variant: "success",
      });
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
                    (child) => child.name !== variables.revision
                  ),
                }
              : {}),
          };
        }
      );
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

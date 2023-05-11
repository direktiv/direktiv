import { useMutation, useQueryClient } from "@tanstack/react-query";

import { TagCreatedSchema } from "../schema";
import { apiFactory } from "../../utils";
import { forceLeadingSlash } from "../utils";
import { treeKeys } from "..";
import { useApiKey } from "../../../util/store/apiKey";
import { useNamespace } from "../../../util/store/namespace";
import { useToast } from "../../../design/Toast";

const createTag = apiFactory({
  pathFn: ({
    namespace,
    path,
    ref,
  }: {
    namespace: string;
    path: string;
    ref: string;
  }) =>
    `/api/namespaces/${namespace}/tree${forceLeadingSlash(
      path
    )}?op=tag&ref=${ref}`,
  method: "POST",
  schema: TagCreatedSchema,
});

export const useCreateTag = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutation({
    mutationFn: ({
      path,
      ref,
      tag,
    }: {
      path: string;
      ref: string;
      tag: string;
    }) =>
      createTag({
        apiKey: apiKey ?? undefined,
        params: { tag },
        pathParams: {
          namespace: namespace,
          path,
          ref,
        },
      }),
    onSuccess: (_, variables) => {
      // update tags and revisions cache. the order ins not predictable, so no way to update the cache ourselves)
      queryClient.invalidateQueries(
        treeKeys.tagsList(namespace, {
          apiKey: apiKey ?? undefined,
          path: variables.path,
        })
      );
      queryClient.invalidateQueries(
        treeKeys.revisionsList(namespace, {
          apiKey: apiKey ?? undefined,
          path: variables.path,
        })
      );
      toast({
        title: "Tag created",
        description: `Tag ${variables.tag} was created`,
        variant: "success",
      });
    },
    onError: () => {
      toast({
        title: "An error occurred",
        description: "could not create tag 😢",
        variant: "error",
      });
    },
  });
};

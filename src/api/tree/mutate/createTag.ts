import { TagCreatedSchema } from "../schema/node";
import { apiFactory } from "~/api/apiFactory";
import { forceLeadingSlash } from "../utils";
import { treeKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const createTag = apiFactory({
  url: ({
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

export const useCreateTag = ({ onSuccess }: { onSuccess?: () => void }) => {
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
        payload: { tag },
        urlParams: {
          namespace,
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
        title: t("api.tree.mutate.createTag.success.title"),
        description: t("api.tree.mutate.createTag.success.description", {
          name: variables.tag,
        }),
        variant: "success",
      });
      onSuccess?.();
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.tree.mutate.createTag.error.description"),
        variant: "error",
      });
    },
  });
};

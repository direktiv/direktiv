import { ToggleLiveSchema } from "../schema/node";
import { apiFactory } from "~/api/apiFactory";
import { forceLeadingSlash } from "../utils";
import { treeKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const toggleLive = apiFactory({
  url: ({ namespace, path }: { namespace: string; path: string }) =>
    `/api/namespaces/${namespace}/tree${forceLeadingSlash(path)}?op=toggle`,
  method: "POST",
  schema: ToggleLiveSchema,
});

export const useToggleLive = ({
  onSuccess,
}: {
  onSuccess?: (data: null) => void;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const { t } = useTranslation();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  const mutationFn = ({ path, value }: { path: string; value: boolean }) =>
    toggleLive({
      apiKey: apiKey ?? undefined,
      payload: { live: value },
      urlParams: {
        namespace,
        path,
      },
    });

  return useMutationWithPermissions({
    mutationFn,
    onSuccess: (data, variables) => {
      queryClient.invalidateQueries(
        treeKeys.router(namespace, {
          apiKey: apiKey ?? undefined,
          path: variables.path,
        })
      );
      onSuccess?.(data);
      const statusKey = variables.value ? "activated" : "deactivated";
      toast({
        title: t("api.tree.mutate.toggleLive.success.title"),
        description: t(
          `api.tree.mutate.toggleLive.success.description.${statusKey}`,
          {
            workflow: variables.path,
            status: variables.value,
          }
        ),
        variant: "success",
      });
    },
  });
};

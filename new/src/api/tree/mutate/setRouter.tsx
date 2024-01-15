import { Trans, useTranslation } from "react-i18next";
import { forceLeadingSlash, sortByRef } from "../utils";

import { RouterSchema } from "../schema/node";
import type { RouterSchemaType } from "../schema/node";
import { apiFactory } from "~/api/apiFactory";
import { treeKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";

const setRouter = apiFactory({
  url: ({ namespace, path }: { namespace: string; path: string }) =>
    `/api/namespaces/${namespace}/tree${forceLeadingSlash(
      path
    )}?op=edit-router`,
  method: "POST",
  schema: RouterSchema,
});

export const useSetRouter = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const { t } = useTranslation();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutationWithPermissions({
    mutationFn: ({
      path,
      routeA,
      routeB,
    }: {
      path: string;
      routeA: { ref: string; weight: number };
      routeB: { ref: string; weight: number };
    }) =>
      setRouter({
        apiKey: apiKey ?? undefined,
        payload: {
          route: [routeA, routeB],
          live: true,
        },
        urlParams: {
          namespace,
          path,
        },
      }),
    onSuccess(data, variables) {
      queryClient.setQueryData<RouterSchemaType>(
        treeKeys.router(namespace, {
          apiKey: apiKey ?? undefined,
          path: variables.path,
        }),
        () => ({
          ...data,
          // TODO: waiting for DIR-576 to get fixed
          routes: [...data.routes.sort(sortByRef)],
        })
      );
      toast({
        title: t("api.tree.mutate.setRouter.success.title"),
        description: (
          <Trans
            i18nKey="api.tree.mutate.setRouter.success.description"
            values={{
              aName: variables.routeA.ref,
              bName: variables.routeB.ref,
              aWeight: variables.routeA.weight,
              bWeight: variables.routeB.weight,
            }}
          />
        ),

        variant: "success",
      });
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.tree.mutate.setRouter.error"),
        variant: "error",
      });
    },
  });
};

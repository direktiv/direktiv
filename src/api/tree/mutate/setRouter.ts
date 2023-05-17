import { forceLeadingSlash, sortByRef } from "../utils";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { RouterSchema } from "../schema";
import type { RouterSchemaType } from "../schema";
import { apiFactory } from "~/api/utils";
import { treeKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
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
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutation({
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
          namespace: namespace,
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
        title: "Restored workflow",
        description: `The latest revision was restored`,
        variant: "success",
      });
    },
    onError: () => {
      toast({
        title: "An error occurred",
        description: "could not revert workflow 😢",
        variant: "error",
      });
    },
  });
};

import { QueryFunctionContext, useQuery } from "@tanstack/react-query";

import { PermissionKeysSchema } from "../schema";
import { permissionKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { z } from "zod";

// const getPermissionKeys = apiFactory({
//   url: ({ baseUrl }: { baseUrl?: string }) =>
//     `${baseUrl ?? ""}/api/v2/info/permissions`,
//   method: "GET",
//   schema: PermissionKeysSchema,
// });

// TODO: remove this mock
const getPermissionKeys = (_params: {
  apiKey?: string;
  urlParams: object;
}): Promise<z.infer<typeof PermissionKeysSchema>> =>
  new Promise((resolve) => {
    setTimeout(() => {
      resolve(
        [
          "opaManage",
          "variablesView",
          "registriesManage",
          "explorerManage",
          "registriesView",
          "nsconfigView",
          "eventsSend",
          "instancesView",
          "secretsView",
          "secretsManage",
          "servicesView",
          "servicesManage",
          "instancesManage",
          "explorerView",
          "workflowView",
          "workflowManage",
          "variablesManage",
          "nsconfigManage",
          "deleteNamespace",
          "eventsView",
          "workflowExecute",
          "workflowStore",
          "permissionsView",
          "permissionsManage",
          "opaView",
          "eventsManage",
        ].sort()
      );
    }, 500);
  });

const fetchpermissionKeys = async ({
  queryKey: [{ apiKey }],
}: QueryFunctionContext<ReturnType<(typeof permissionKeys)["get"]>>) =>
  getPermissionKeys({
    apiKey,
    urlParams: {},
  });

export const usePermissionKeys = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: permissionKeys.get({ apiKey: apiKey ?? undefined }),
    queryFn: fetchpermissionKeys,
    staleTime: Infinity, // this is a long lived static list, no refetch needed until the page is refreshed
  });
};

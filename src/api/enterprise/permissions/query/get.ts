import { QueryFunctionContext, useQuery } from "@tanstack/react-query";

import { PermissionKeysSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { permissionKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { z } from "zod";

// TODO: remove the line below and delete the mock function
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const getPermissionKeys = apiFactory({
  url: ({ baseUrl }: { baseUrl?: string }) =>
    `${baseUrl ?? ""}/api/v2/info/permissions`,
  method: "GET",
  schema: PermissionKeysSchema,
});

const getPermissionsKeysMock = (_params: {
  apiKey?: string;
}): Promise<z.infer<typeof PermissionKeysSchema>> =>
  new Promise((resolve) => {
    setTimeout(() => {
      resolve([
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
      ]);
    }, 500);
  });

const fetchpermissionKeys = async ({
  queryKey: [{ apiKey }],
}: QueryFunctionContext<ReturnType<(typeof permissionKeys)["get"]>>) =>
  getPermissionsKeysMock({
    apiKey,
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

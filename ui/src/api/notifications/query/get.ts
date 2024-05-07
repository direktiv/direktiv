import { NotificationListSchema } from "../schema";
import type { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "../../apiFactory";
import { notificationKeys } from "..";
import { useApiKey } from "../../../util/store/apiKey";
import { useNamespace } from "../../../util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

const getNotifications = apiFactory({
  url: ({ namespace }: { namespace: string }) =>
    `/api/v2/namespaces/${namespace}/notifications`,
  method: "GET",
  schema: NotificationListSchema,
});

const fetchNotifications = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<
  ReturnType<(typeof notificationKeys)["notifications"]>
>) =>
  getNotifications({
    apiKey,
    urlParams: { namespace },
  });

export const useNotifications = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: notificationKeys.notifications(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchNotifications,
    enabled: !!namespace,
  });
};

import {
  EventCreatedSchema,
  EventCreatedSchemaType,
  NewEventSchemaType,
} from "../schema";

import { apiFactory } from "~/api/apiFactory";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

/**
 * TODO: apiFactory<NewEventSchemaType["body"]> should be the correct type
 *  but the but e2e/utils/events.ts is sending something different.
 */
export const sendEvent = apiFactory({
  url: ({ baseUrl, namespace }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/namespaces/${namespace}/broadcast`,
  method: "POST",
  schema: EventCreatedSchema,
});

export const useSendEvent = ({
  onSuccess,
}: {
  onSuccess?: (data: EventCreatedSchemaType) => void;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const { t } = useTranslation();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  const mutationFn = (data: NewEventSchemaType) =>
    sendEvent({
      apiKey: apiKey ?? undefined,
      payload: data.body,
      headers: {
        "content-type": "application/cloudevents+json",
      },
      urlParams: {
        namespace,
      },
    });

  return useMutationWithPermissions({
    mutationFn,
    onSuccess: () => {
      /* Cache for scope: events-list is not invalidated here because in the
       * current use case, new events come in via streaming and additional
       * complexity would be required to invalidate cache. If cache invalidation
       * becomes relevant, it should probably be optional so no refetching
       * occurs in views that use streaming.
       * */
      toast({
        title: t("api.events.mutate.sendEvent.success.title"),
        description: t("api.events.mutate.sendEvent.success.description"),
        variant: "success",
      });
      onSuccess?.(null);
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.events.mutate.sendEvent.error.description"),
        variant: "error",
      });
    },
  });
};

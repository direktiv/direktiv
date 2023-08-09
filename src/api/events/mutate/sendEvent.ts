import {
  EventCreatedSchema,
  EventCreatedSchemaType,
  NewEventSchemaType,
} from "../schema";

import { apiFactory } from "~/api/apiFactory";
import { useApiKey } from "~/util/store/apiKey";
import { useMutation } from "@tanstack/react-query";
import { useNamespace } from "~/util/store/namespace";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

export const sendEvent = apiFactory({
  url: ({ baseUrl, namespace }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/namespaces/${namespace}/broadcast`,
  method: "POST",
  schema: EventCreatedSchema,
});

export const useSendEvent = ({
  onSuccess,
}: {
  onSuccess?: (registry: EventCreatedSchemaType) => void;
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

  return useMutation({
    mutationFn,
    // no cache invalidation because events are updated via streaming
    onSuccess: () => {
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

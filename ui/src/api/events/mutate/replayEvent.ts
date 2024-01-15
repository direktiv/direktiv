import { EventCreatedSchema, EventCreatedSchemaType } from "../schema";

import { apiFactory } from "~/api/apiFactory";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

export const replayEvent = apiFactory({
  url: ({
    baseUrl,
    namespace,
    id,
  }: {
    baseUrl?: string;
    namespace: string;
    id: string;
  }) => `${baseUrl ?? ""}/api/namespaces/${namespace}/events/${id}/replay`,

  method: "POST",
  schema: EventCreatedSchema,
});

export const useReplayEvent = ({
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

  const mutationFn = (id: string) =>
    replayEvent({
      apiKey: apiKey ?? undefined,
      urlParams: {
        id,
        namespace,
      },
    });

  return useMutationWithPermissions({
    mutationFn,
    onSuccess: () => {
      toast({
        title: t("api.events.mutate.replayEvent.success.title"),
        description: t("api.events.mutate.replayEvent.success.description"),
        variant: "success",
      });
      onSuccess?.(null);
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.events.mutate.replayEvent.error.description"),
        variant: "error",
      });
    },
  });
};

import {
  VarCreatedUpdatedSchema,
  VarFormCreateEditSchemaType,
} from "../schema";

import { apiFactory } from "~/api/apiFactory";
import { forceLeadingSlash } from "~/api/files/utils";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";
import { varKeys } from "..";

export const createVar = apiFactory<VarFormCreateEditSchemaType>({
  url: ({ baseUrl, namespace }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/variables`,
  method: "POST",
  schema: VarCreatedUpdatedSchema,
});

export const useCreateVar = ({
  onSuccess,
}: {
  onSuccess?: () => void;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  const mutationFn = (payload: VarFormCreateEditSchemaType) =>
    createVar({
      apiKey: apiKey ?? undefined,
      payload: {
        ...payload,
        workflowPath: payload.workflowPath
          ? forceLeadingSlash(payload.workflowPath)
          : undefined,
      },
      urlParams: {
        namespace,
      },
    });

  return useMutationWithPermissions({
    mutationFn,
    onSuccess: (data) => {
      queryClient.invalidateQueries(
        varKeys.varList(namespace, {
          apiKey: apiKey ?? undefined,
          workflowPath:
            data.data.type === "workflow-variable"
              ? data.data.reference
              : undefined,
        })
      );
      toast({
        title: t("api.variables.mutate.updateVariable.success.title"),
        description: t(
          "api.variables.mutate.updateVariable.success.description",
          { name: data.data.name }
        ),
        variant: "success",
      });
      onSuccess?.();
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.variables.mutate.updateVariable.error.description"),
        variant: "error",
      });
    },
  });
};

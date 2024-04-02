import {
  VarCreatedSchema,
  VarCreatedSchemaType,
  VarFormSchemaType,
} from "../schema";

import { apiFactory } from "~/api/apiFactory";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";
import { varKeys } from "..";

export const createVar = apiFactory({
  url: ({ baseUrl, namespace }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/variables`,
  method: "PUT",
  schema: VarCreatedSchema,
});

export const useCreateVar = ({
  onSuccess,
}: {
  onSuccess?: (data: VarCreatedSchemaType) => void;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  const mutationFn = (payload: VarFormSchemaType) =>
    createVar({
      apiKey: apiKey ?? undefined,
      payload,
      urlParams: {
        namespace,
      },
    });

  return useMutationWithPermissions({
    mutationFn,
    onSuccess: (data, variables) => {
      // Two cache invalidations are needed due to the current API,
      // which uses the same endpoint for creating and editing.
      // varContent needs a refresh after editing, varList needs a
      // refresh after creating (the variable's content is not
      // included in the list)
      queryClient.invalidateQueries(
        varKeys.varContent(namespace, {
          apiKey: apiKey ?? undefined,
          name: variables.name,
        })
      );
      queryClient.invalidateQueries(
        varKeys.varList(namespace, {
          apiKey: apiKey ?? undefined,
        })
      );
      toast({
        title: t("api.variables.mutate.updateVariable.success.title"),
        description: t(
          "api.variables.mutate.updateVariable.success.description",
          { name: data.key }
        ),
        variant: "success",
      });
      onSuccess?.(data);
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

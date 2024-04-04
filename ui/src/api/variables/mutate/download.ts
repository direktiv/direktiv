import { VarContentType, getVariableContent } from "../query/details";

import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

export const useDownloadVar = ({
  onSuccess,
}: {
  onSuccess?: (varContent: VarContentType, name: string) => void;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const { t } = useTranslation();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  const mutationFn = (variableID: string) =>
    getVariableContent({
      apiKey: apiKey ?? undefined,
      urlParams: {
        namespace,
        variableID,
      },
    });

  return useMutationWithPermissions({
    mutationFn,
    onSuccess: (data, name) => {
      onSuccess?.(data, name);
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t(
          "api.variables.mutate.downloadVariable.error.description"
        ),
        variant: "error",
      });
    },
  });
};

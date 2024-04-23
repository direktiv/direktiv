import { VarDetailsType, getVariableDetails } from "../query/details";

import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

export const useDownloadVar = ({
  onSuccess,
}: {
  onSuccess?: (variable: VarDetailsType) => void;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const { t } = useTranslation();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  const mutationFn = (id: string) =>
    getVariableDetails({
      apiKey: apiKey ?? undefined,
      urlParams: {
        namespace,
        id,
      },
    });

  return useMutationWithPermissions({
    mutationFn,
    onSuccess: (data) => {
      onSuccess?.(data);
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

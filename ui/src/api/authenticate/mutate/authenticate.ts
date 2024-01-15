import { checkApiKeyAgainstServer } from "..";
import { toast } from "~/design/Toast";
import { useMutation } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";

export const useAuthenticate = ({
  onSuccess,
}: {
  onSuccess?: (isKeyCorrect: boolean, apiKey: string) => void;
} = {}) => {
  const { t } = useTranslation();
  return useMutation({
    mutationFn: (apiKey: string) => checkApiKeyAgainstServer(apiKey),
    onSuccess: (isKeyCorrect, apiKey) => {
      onSuccess?.(isKeyCorrect, apiKey);
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t(
          "api.authentication.mutate.authenticate.error.description"
        ),
        variant: "error",
      });
    },
  });
};

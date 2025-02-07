import { TokenDeletedSchema, TokenListSchemaType } from "../schema";

import { apiFactory } from "~/api/apiFactory";
import { tokenKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const updateCache = (
  oldData: TokenListSchemaType | undefined,
  tokenName: Parameters<ReturnType<typeof useDeleteToken>["mutate"]>[0]
) => {
  if (!oldData) return undefined;
  const remainingTokens = oldData.data.filter(
    (token) => token.name !== tokenName
  );
  return {
    ...oldData,
    tokens: remainingTokens,
  };
};

const deleteToken = apiFactory({
  url: ({ namespace, tokenName }: { namespace: string; tokenName: string }) =>
    `/api/v2/namespaces/${namespace}/api_tokens/${tokenName}`,
  method: "DELETE",
  schema: TokenDeletedSchema,
});

export const useDeleteToken = ({
  onSuccess,
}: { onSuccess?: () => void } = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutationWithPermissions({
    mutationFn: (tokenName: string) =>
      deleteToken({
        apiKey: apiKey ?? undefined,
        urlParams: { tokenName, namespace },
      }),
    onSuccess() {
      queryClient.invalidateQueries({
        queryKey: tokenKeys.apiTokens(namespace, {
          apiKey: apiKey ?? undefined,
        }),
      });
      toast({
        title: t("api.tokens.mutate.deleteToken.success.title"),
        description: t("api.tokens.mutate.deleteToken.success.description"),
        variant: "success",
      });
      onSuccess?.();
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.tokens.mutate.deleteToken.error.description"),
        variant: "error",
      });
    },
  });
};

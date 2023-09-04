import {
  TokenDeletedSchema,
  TokenListSchemaType,
  TokenSchemaType,
} from "../schema";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { apiFactory } from "~/api/apiFactory";
import { tokenKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";
import { z } from "zod";

const updateCache = (
  oldData: TokenListSchemaType | undefined,
  variables: Parameters<ReturnType<typeof useDeleteToken>["mutate"]>[0]
) => {
  if (!oldData) return undefined;
  const remainingTokens = oldData.tokens.filter(
    (token) => token.id !== variables.id
  );
  return {
    ...oldData,
    tokens: remainingTokens,
  };
};

// TODO: remove the line below and delete the mock function
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const deleteToken = apiFactory({
  url: ({ namespace, tokenId }: { namespace: string; tokenId: string }) =>
    `/api/v2/namespaces/${namespace}/tokens/${tokenId}`,
  method: "DELETE",
  schema: TokenDeletedSchema,
});

const deleteTockenMock = (_params: {
  apiKey?: string;
  urlParams: { namespace: string; tokenId: string };
}): Promise<z.infer<typeof TokenDeletedSchema>> =>
  new Promise((resolve) => {
    setTimeout(() => {
      resolve(null);
    }, 500);
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

  return useMutation({
    mutationFn: (token: TokenSchemaType) =>
      deleteTockenMock({
        apiKey: apiKey ?? undefined,
        urlParams: {
          tokenId: token.id,
          namespace,
        },
      }),
    onSuccess(_, variables) {
      queryClient.setQueryData<TokenListSchemaType>(
        tokenKeys.tokenList(namespace, {
          apiKey: apiKey ?? undefined,
        }),
        (oldData) => updateCache(oldData, variables)
      );
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

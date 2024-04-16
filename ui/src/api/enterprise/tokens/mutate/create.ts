import { TokenCreatedSchema, TokenFormSchemaType } from "../schema";

import { apiFactory } from "~/api/apiFactory";
import { tokenKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const createToken = apiFactory({
  url: ({ namespace, baseUrl }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/tokens`,
  method: "POST",
  schema: TokenCreatedSchema,
});

type ResolvedCreateToken = Awaited<ReturnType<typeof createToken>>;

export const useCreateToken = ({
  onSuccess,
}: { onSuccess?: (data: ResolvedCreateToken) => void } = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutationWithPermissions({
    mutationFn: (tokenFormProps: TokenFormSchemaType) =>
      createToken({
        apiKey: apiKey ?? undefined,
        urlParams: {
          namespace,
        },
        payload: tokenFormProps,
      }),
    onSuccess(data) {
      queryClient.invalidateQueries({
        queryKey: tokenKeys.tokenList(namespace, {
          apiKey: apiKey ?? undefined,
        }),
      });
      onSuccess?.(data);
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.tokens.mutate.createToken.error.description"),
        variant: "error",
      });
    },
  });
};

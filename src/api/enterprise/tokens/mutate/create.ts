import { TokenCreatedSchema, TokenFormSchemaType } from "../schema";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { faker } from "@faker-js/faker";
import { tokenKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";
import { z } from "zod";

// const createToken = apiFactory({
//   url: ({ namespace, baseUrl }: { baseUrl?: string; namespace: string }) =>
//     `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/tokens`,
//   method: "POST",
//   schema: TokenCreatedSchema,
// });

// TODO: remove this mock
const createToken = (_params: {
  apiKey?: string;
  urlParams: { namespace: string };
  payload: TokenFormSchemaType;
}): Promise<z.infer<typeof TokenCreatedSchema>> =>
  new Promise((resolve) => {
    setTimeout(
      () =>
        resolve({
          id: faker.datatype.uuid(),
          token: faker.datatype.uuid(),
        }),
      500
    );
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

  return useMutation({
    mutationFn: (tokenFormProps: TokenFormSchemaType) =>
      createToken({
        apiKey: apiKey ?? undefined,
        urlParams: {
          namespace,
        },
        payload: tokenFormProps,
      }),
    onSuccess(data) {
      queryClient.invalidateQueries(
        tokenKeys.tokenList(namespace, {
          apiKey: apiKey ?? undefined,
        })
      );
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

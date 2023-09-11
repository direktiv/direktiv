import { ResponseParser, apiFactory } from "~/api/apiFactory";

import { VarDownloadSchema } from "../schema";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const blobResponseParser: ResponseParser = async ({ res, schema }) => {
  const varBlob = await res.blob();
  const headers = Object.fromEntries(res.headers);
  return schema.parse({ blob: varBlob, headers });
};

export const getVarBlob = apiFactory({
  url: ({ namespace, name }: { namespace: string; name: string }) =>
    `/api/namespaces/${namespace}/vars/${name}`,
  method: "GET",
  schema: VarDownloadSchema,
  responseParser: blobResponseParser,
});

type VarContentType = Awaited<ReturnType<typeof getVarBlob>>;

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

  const mutationFn = (name: string) =>
    getVarBlob({
      apiKey: apiKey ?? undefined,
      urlParams: {
        namespace,
        name,
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

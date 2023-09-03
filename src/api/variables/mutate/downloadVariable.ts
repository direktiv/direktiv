import { ResponseParser, apiFactory } from "~/api/apiFactory";

import { VarContentSchema } from "../schema";
import { useApiKey } from "~/util/store/apiKey";
import { useMutation } from "@tanstack/react-query";
import { useNamespace } from "~/util/store/namespace";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const variableResponseParser: ResponseParser = async ({ res, schema }) => {
  // Different from the default responseParser, for varibles we
  // don't want to try to parse the response as JSON because
  // it will always be treated as text to be displayed or edited
  // (even if it is JSON).
  const textResult = await res.text();
  if (!textResult) return schema.parse(null);

  // If response is not null, return it as 'body',
  // and also add the response headers (content type is needed later)
  // This is a workaround, in the new API we should return the
  // content type as part of the JSON response.
  const headers = Object.fromEntries(res.headers);
  return schema.parse({ body: textResult, headers });
};

export const getVarBlob = apiFactory({
  url: ({ namespace, name }: { namespace: string; name: string }) =>
    `/api/namespaces/${namespace}/vars/${name}`,
  method: "GET",
  schema: VarContentSchema,
  responseParser: variableResponseParser,
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

  return useMutation({
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

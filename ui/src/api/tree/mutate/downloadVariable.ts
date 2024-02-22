import { ResponseParser, apiFactory } from "~/api/apiFactory";

import { WorkflowVariableDownloadSchema } from "../schema/workflowVariable";
import { forceLeadingSlash } from "~/api/files/utils";
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
  url: ({
    namespace,
    name,
    path,
  }: {
    namespace: string;
    name: string;
    path: string;
  }) =>
    `/api/namespaces/${namespace}/tree${forceLeadingSlash(
      path
    )}?op=var&var=${name}`,
  method: "GET",
  schema: WorkflowVariableDownloadSchema,
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

  const mutationFn = ({ name, path }: { name: string; path: string }) =>
    getVarBlob({
      apiKey: apiKey ?? undefined,
      urlParams: {
        namespace,
        name,
        path,
      },
    });

  return useMutationWithPermissions({
    mutationFn,
    onSuccess: (data, { name }) => {
      onSuccess?.(data, name);
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.tree.mutate.downloadVariable.error.description"),
        variant: "error",
      });
    },
  });
};

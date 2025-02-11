import { QueryFunctionContext, useQueryClient } from "@tanstack/react-query";

import { InstanceOutputResponseSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { instanceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

export const getInstanceOutput = apiFactory({
  url: ({
    namespace,
    baseUrl,
    instanceId,
  }: {
    baseUrl?: string;
    namespace: string;
    instanceId: string;
  }) =>
    `${
      baseUrl ?? ""
    }/api/v2/namespaces/${namespace}/instances/${instanceId}/output`,
  method: "GET",
  schema: InstanceOutputResponseSchema,
});

const fetchInstanceOutput = async ({
  queryKey: [{ apiKey, namespace, instanceId }],
}: QueryFunctionContext<ReturnType<(typeof instanceKeys)["instanceOutput"]>>) =>
  getInstanceOutput({
    apiKey,
    urlParams: { namespace, instanceId },
  });

export const useInstanceOutput = ({
  instanceId,
  enabled,
}: {
  instanceId: string;
  enabled?: boolean;
}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: instanceKeys.instanceOutput(namespace, {
      apiKey: apiKey ?? undefined,
      instanceId,
    }),
    queryFn: fetchInstanceOutput,
    enabled,
    select: (data) => data.data,
  });
};

export const getInstanceOutputForPath = apiFactory({
  url: ({
    namespace,
    baseUrl,
    path,
  }: {
    baseUrl?: string;
    namespace: string;
    path: string;
  }) =>
    `${
      baseUrl ?? ""
    }/api/v2/namespaces/${namespace}/instances?path=${path}/&wait=true&output=true`,
  method: "POST",
  schema: InstanceOutputResponseSchema,
});

const fetchInstanceOutputForPath = async ({
  queryKey: [{ apiKey, namespace, path }],
}: QueryFunctionContext<
  ReturnType<(typeof instanceKeys)["instanceOutputForPath"]>
>) =>
  getInstanceOutputForPath({
    apiKey,
    urlParams: { namespace, path },
  });

export const useInstanceOutputForPath = ({
  onSuccess,
}: { onSuccess?: (data: unknown) => void } = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }
  return useMutationWithPermissions({
    mutationFn: (tokenFormProps: any) =>
      getInstanceOutputForPath({
        apiKey: apiKey ?? undefined,
        urlParams: {
          namespace,
          path,
        },
        payload: tokenFormProps,
      }),
    onSuccess(data, { description }) {
      queryClient.invalidateQueries({
        queryKey: groupKeys.groupList(namespace, {
          apiKey: apiKey ?? undefined,
        }),
      });
      toast({
        title: t("api.groups.mutate.createGroup.success.title"),
        description: t("api.groups.mutate.createGroup.success.description", {
          name: description,
        }),
        variant: "success",
      });
      onSuccess?.(data);
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.groups.mutate.createGroup.error.description"),
        variant: "error",
      });
    },
  });
};

import { OpenapiSpecificationSchema } from "../schema";
import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { gatewayKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

export const getDocumentation = apiFactory({
  url: ({ baseUrl, namespace }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/gateway/info?expand=true`,
  method: "GET",
  schema: OpenapiSpecificationSchema,
});

const fetchDocumentation = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof gatewayKeys)["documentation"]>>) =>
  getDocumentation({
    apiKey,
    urlParams: { namespace },
  });

export const useDocumentation = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: gatewayKeys.documentation(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchDocumentation,
    enabled: !!namespace,
  });
};

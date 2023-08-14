import { QueryFunctionContext, useQuery } from "@tanstack/react-query";

import { ServicesListSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { serviceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";

export const getServices = apiFactory({
  url: ({ namespace, baseUrl }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/functions/namespaces/${namespace}`,
  method: "GET",
  schema: ServicesListSchema,
});

const fetchServices = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof serviceKeys)["servicesList"]>>) =>
  getServices({
    apiKey,
    urlParams: { namespace },
  });

export const useServices = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: serviceKeys.servicesList(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchServices,
    enabled: !!namespace,
  });
};

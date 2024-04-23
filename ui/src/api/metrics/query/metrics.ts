import { QueryFunctionContext, useQuery } from "@tanstack/react-query";

import { MetricsResponseSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { forceLeadingSlash } from "~/api/files/utils";
import { metricsKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";

const getMetrics = apiFactory({
  url: ({ namespace, path }: { namespace: string; path?: string }) =>
    `/api/v2/namespaces/${namespace}/metrics/instances?workflowPath=${forceLeadingSlash(
      path
    )}`,
  method: "GET",
  schema: MetricsResponseSchema,
});

const fetchMetrics = async ({
  queryKey: [{ apiKey, namespace, path }],
}: QueryFunctionContext<ReturnType<(typeof metricsKeys)["metrics"]>>) =>
  getMetrics({
    apiKey,
    urlParams: {
      namespace,
      path,
    },
  });

export const useMetrics = ({ path }: { path?: string }) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: metricsKeys.metrics(namespace, {
      apiKey: apiKey ?? undefined,
      path,
    }),
    queryFn: fetchMetrics,
    enabled: !!namespace,
  });
};

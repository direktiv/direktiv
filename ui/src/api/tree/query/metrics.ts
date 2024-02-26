import { QueryFunctionContext, useQuery } from "@tanstack/react-query";

import { MetricsListSchema } from "../schema/metrics";
import { apiFactory } from "~/api/apiFactory";
import { forceLeadingSlash } from "~/api/files/utils";
import { treeKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";

type MetricsType = "successful" | "failed";

const getMetrics = apiFactory({
  url: ({
    namespace,
    path,
    type,
  }: {
    namespace: string;
    path?: string;
    type: MetricsType;
  }) =>
    `/api/namespaces/${namespace}/tree${forceLeadingSlash(
      path
    )}?op=metrics-${type}`,
  method: "GET",
  schema: MetricsListSchema,
});

const fetchMetrics = async ({
  queryKey: [{ apiKey, namespace, path, type }],
}: QueryFunctionContext<ReturnType<(typeof treeKeys)["metrics"]>>) =>
  getMetrics({
    apiKey,
    urlParams: {
      namespace,
      path,
      type,
    },
  });

export const useMetrics = ({
  path,
  type,
}: {
  path?: string;
  type: MetricsType;
}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: treeKeys.metrics(namespace, {
      apiKey: apiKey ?? undefined,
      path,
      type,
    }),
    queryFn: fetchMetrics,
    enabled: !!namespace,
  });
};

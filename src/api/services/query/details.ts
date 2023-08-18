import { QueryFunctionContext, useQuery } from "@tanstack/react-query";

import { ServicesRevisionListSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { serviceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";

export const getServiceDetails = apiFactory({
  url: ({
    namespace,
    service,
    baseUrl,
  }: {
    baseUrl?: string;
    namespace: string;
    service: string;
  }) =>
    `${
      baseUrl ?? ""
    }/api/functions/namespaces/${namespace}/function/${service}`,
  method: "GET",
  schema: ServicesRevisionListSchema,
});

const fetchServiceDetails = async ({
  queryKey: [{ apiKey, namespace, service }],
}: QueryFunctionContext<ReturnType<(typeof serviceKeys)["serviceDetail"]>>) =>
  getServiceDetails({
    apiKey,
    urlParams: { namespace, service },
  }).then((res) => ({
    // revisions must be sorted by creation date, to figure out the latest revision
    ...res,
    revisions: (res.revisions ?? []).sort((a, b) => {
      if (a.revision > b.revision) {
        return -1;
      }
      if (a.revision < b.revision) {
        return 1;
      }
      return 0;
    }),
  }));

export const useServiceDetails = ({ service }: { service: string }) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: serviceKeys.serviceDetail(namespace, {
      apiKey: apiKey ?? undefined,
      service,
    }),
    queryFn: fetchServiceDetails,
    enabled: !!namespace,
  });
};

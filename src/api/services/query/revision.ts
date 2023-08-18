import { ServicesRevisionListSchemaType } from "../schema";
import { serviceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useQuery } from "@tanstack/react-query";

/**
 * The queryFn of this hook will never return any data because we only have a
 * streaming endoint for this data. This hook is only used to subscribe to the
 * correct cache key. Data for this key will be added by a streaming subscriber
 */
export const useServiceRevision = ({
  service,
  revision,
}: {
  service: string;
  revision: string;
}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: serviceKeys.serviceRevisionDetail(namespace, {
      apiKey: apiKey ?? undefined,
      service,
      revision,
    }),
    queryFn: (): ServicesRevisionListSchemaType | undefined => undefined,
    enabled: !!namespace,
  });
};

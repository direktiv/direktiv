import { InstancesDetailSchema, InstancesDetailSchemaType } from "../schema";
import {
  QueryFunctionContext,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";

import { apiFactory } from "~/api/apiFactory";
import { instanceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useStreaming } from "~/api/streaming";

export type FiltersObj = {
  QUERY?: { type: "MATCH"; workflowName?: string; stateName?: string };
};

export const getFilterQuery = (filters: FiltersObj) => {
  let query = "";
  const filterFields = Object.keys(filters) as Array<keyof FiltersObj>;

  filterFields.forEach((field) => {
    const filterItem = filters[field];
    // without the guard, TS thinks filterItem may be undefined
    if (!filterItem) {
      return console.error("filterItem is not defined");
    }

    if (field === "QUERY") {
      const workflowName = filterItem?.workflowName ?? "";
      const stateName = filterItem?.stateName ?? "";
      query = query.concat(
        `&filter.field=${field}&filter.type=${filterItem.type}&filter.val=${workflowName}::${stateName}::`
      );
    }
  });

  return query;
};

const getUrl = ({
  namespace,
  baseUrl,
  instanceId,
  filters,
}: {
  baseUrl?: string;
  namespace: string;
  instanceId: string;
  filters?: FiltersObj;
}) => {
  let url = `${
    baseUrl ?? ""
  }/api/namespaces/${namespace}/instances/${instanceId}`;

  if (filters) {
    url = url.concat(`?${filters}`);
  }

  return url;
};

export const getInstanceDetails = apiFactory({
  url: getUrl,
  method: "GET",
  schema: InstancesDetailSchema,
});

const fetchInstanceDetails = async ({
  queryKey: [{ apiKey, namespace, instanceId, filters }],
}: QueryFunctionContext<ReturnType<(typeof instanceKeys)["instanceDetail"]>>) =>
  getInstanceDetails({
    apiKey,
    urlParams: { namespace, instanceId, filters },
  });

export const useInstanceDetailsStream = ({
  instanceId,
  filters,
  enabled = true,
}: {
  instanceId: string;
  enabled?: boolean;
  filters: FiltersObj;
}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useStreaming({
    url: `/api/namespaces/${namespace}/instances/${instanceId}`,
    enabled,
    schema: InstancesDetailSchema,
    onMessage: (msg) => {
      queryClient.setQueryData<InstancesDetailSchemaType>(
        instanceKeys.instanceDetail(namespace, {
          apiKey: apiKey ?? undefined,
          instanceId,
          filters,
        }),
        () => msg
      );
    },
  });
};

export const useInstanceDetails = ({
  instanceId,
  filters,
}: {
  instanceId: string;
  filters: FiltersObj;
}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: instanceKeys.instanceDetail(namespace, {
      apiKey: apiKey ?? undefined,
      instanceId,
      filters,
    }),
    queryFn: fetchInstanceDetails,
    enabled: !!namespace,
  });
};

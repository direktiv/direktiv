import { QueryFunctionContext, useQuery } from "@tanstack/react-query";

import { InstancesListSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { instanceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";

export const triggerValues = ["api", "cloudevent", "instance", "cron"] as const;

export const statusValues = [
  "pending",
  "complete",
  "cancelled",
  "failed",
] as const;

export type TriggerValue = (typeof triggerValues)[number];
export type StatusValue = (typeof statusValues)[number];
export type FiltersObj = {
  AS?: { type: "CONTAINS"; value: string };
  STATUS?: {
    type: "MATCH";
    value: StatusValue;
  };
  TRIGGER?: {
    type: "MATCH";
    value: TriggerValue;
  };
  AFTER?: { type: "AFTER"; value: Date };
  BEFORE?: { type: "BEFORE"; value: Date };
};

export const getFilterQuery = (filters: FiltersObj) => {
  let query = "";
  const filterFields = Object.keys(filters) as Array<keyof FiltersObj>;

  filterFields.forEach((field) => {
    const filterItem = filters[field];

    // Without the guard, TS thinks filterItem may be undefined
    if (!filterItem) {
      return console.error("filterItem is not defined");
    }

    let queryField: string;
    let queryValue: string;

    if (field === "AFTER" || field === "BEFORE") {
      const date = filters[field]?.value;
      if (!date) {
        throw new Error("date is not defined in date filter");
      }
      queryField = "CREATED";
      queryValue = date.toISOString();
    } else {
      const value = filters[field]?.value;
      if (!value) {
        throw new Error("filter value is not defined");
      }
      queryField = field;
      queryValue = value;
    }

    query = query.concat(
      `&filter.field=${queryField}&filter.type=${filterItem.type}&filter.val=${queryValue}`
    );
  });

  return query;
};

const getUrl = ({
  namespace,
  baseUrl,
  limit,
  offset,
  filters,
}: {
  baseUrl?: string;
  namespace: string;
  limit: number;
  offset: number;
  filters?: FiltersObj;
}) => {
  let url = `${
    baseUrl ?? ""
  }/api/namespaces/${namespace}/instances?limit=${limit}&offset=${offset}`;
  if (filters) {
    url = url.concat(getFilterQuery(filters));
  }
  return url;
};

export const getInstances = apiFactory({
  url: getUrl,
  method: "GET",
  schema: InstancesListSchema,
});

const fetchInstances = async ({
  queryKey: [{ apiKey, namespace, limit, offset, filters }],
}: QueryFunctionContext<ReturnType<(typeof instanceKeys)["instancesList"]>>) =>
  getInstances({
    apiKey,
    urlParams: { namespace, limit, offset, filters },
  });

export const useInstances = ({
  limit,
  offset,
  filters,
}: {
  limit: number;
  offset: number;
  filters: FiltersObj;
}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: instanceKeys.instancesList(namespace, {
      apiKey: apiKey ?? undefined,
      limit,
      offset,
      filters,
    }),
    queryFn: fetchInstances,
    enabled: !!namespace,
  });
};

import { QueryFunctionContext, useQuery } from "@tanstack/react-query";

import { InstancesListSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { instanceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";

export type FiltersObj = {
  AS?: { type: "CONTAINS"; value: string };
  STATUS?: {
    type: "MATCH";
    value: "pending" | "complete" | "cancelled" | "failed";
  };
  TRIGGER?: {
    type: "MATCH";
    value: "api" | "cloudevent" | "instance" | "cron";
  };
  // TODO: use Date type (but display components need a string value)
  AFTER?: { type: "AFTER"; value: Date };
  BEFORE?: { type: "BEFORE"; value: Date };
};

const getFilterQuery = (filters: FiltersObj) => {
  let query = "";
  const filterFields = Object.keys(filters) as Array<keyof FiltersObj>;

  filterFields.forEach((field) => {
    const filterItem = filters[field];

    // Without the guard, TS thinks filterItem may be undefined
    if (!filterItem) {
      throw new Error("filterItem is not defined");
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

export const getInstances = apiFactory({
  url: ({
    namespace,
    baseUrl,
    limit,
    offset,
    filter,
  }: {
    baseUrl?: string;
    namespace: string;
    limit: number;
    offset: number;
    filter?: string;
  }) =>
    `${
      baseUrl ?? ""
    }/api/namespaces/${namespace}/instances?limit=${limit}&offset=${offset}${
      filter ?? filter
    }`,
  method: "GET",
  schema: InstancesListSchema,
});

const fetchInstances = async ({
  queryKey: [{ apiKey, namespace, limit, offset, filter }],
}: QueryFunctionContext<ReturnType<(typeof instanceKeys)["instancesList"]>>) =>
  getInstances({
    apiKey,
    urlParams: { namespace, limit, offset, filter },
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
      filter: getFilterQuery(filters),
    }),
    queryFn: fetchInstances,
    enabled: !!namespace,
  });
};

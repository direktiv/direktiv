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
  AFTER?: { type: "AFTER"; value: string };
  BEFORE?: { type: "BEFORE"; value: string };
};

const getFilterQuery = (filters: FiltersObj) => {
  let query = "";
  const filterFields = Object.keys(filters) as Array<keyof FiltersObj>;

  filterFields.forEach((field) => {
    const filterItem = filters[field];
    // guard needed because TS thinks filterItem may be undefined
    if (!filterItem) {
      return;
    }

    // FilterFields uses BEFORE and AFTER as distinct types. In the query,
    // the format is filter.field=CREATED&filter.type=BEFORE|AFTER
    const queryField = ["AFTER", "BEFORE"].includes(field) ? "CREATED" : field;

    query = query.concat(
      `&filter.field=${queryField}&filter.type=${filterItem.type}&filter.val=${filterItem.value}`
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

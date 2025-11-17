import { possibleInstanceStatuses, possibleInvokerValues } from "../schema";

export const statusValues = possibleInstanceStatuses;

export const invokerValues = possibleInvokerValues;

type invokerValuesType = typeof possibleInvokerValues | `instance:${string}`;

export type InvokerValue = invokerValuesType[number];
export type StatusValue = (typeof statusValues)[number];

export type FiltersObj = {
  path?: { operator?: "cn"; value: string };
  status?: {
    operator?: "eq";
    value: StatusValue;
  };
  invoker?: {
    operator?: "eq";
    value: InvokerValue;
  };
  createdAtLt?: {
    operator: "lt";
    value: Date;
  };
  createdAtGt?: {
    operator: "gt";
    value: Date;
  };
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

    const queryOperator = filters[field]?.operator ?? undefined;

    if (field === "createdAtLt" || field === "createdAtGt") {
      const date = filters[field]?.value;
      if (!date) {
        throw new Error("date is not defined in date filter");
      }
      queryField = "createdAt";
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
      `&filter[${queryField}]${queryOperator ? `[${queryOperator}]` : ""}=${queryValue}`
    );
  });

  return query;
};

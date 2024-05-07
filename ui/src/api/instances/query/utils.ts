import { possibleInstanceStatuses, possibleTriggerValues } from "../schema";

export const statusValues = possibleInstanceStatuses;

export const triggerValues = possibleTriggerValues;

export type triggerValuesType =
  | typeof possibleTriggerValues
  | `instance:${string}`;

export type TriggerValue = triggerValuesType[number];
export type StatusValue = (typeof statusValues)[number];

export type FiltersObj = {
  AS?: { type: "CONTAINS" | "WORKFLOW"; value: string };
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

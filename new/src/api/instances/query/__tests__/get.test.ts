import { FiltersObj, getFilterQuery } from "../get";
import { describe, expect, test } from "vitest";

const filterByAs: FiltersObj = {
  AS: { type: "CONTAINS", value: "Findme" },
};

const filterByStatus: FiltersObj = {
  STATUS: { type: "MATCH", value: "failed" },
};

const filterByTrigger: FiltersObj = {
  TRIGGER: { type: "MATCH", value: "cloudevent" },
};

const filterByAfter: FiltersObj = {
  AFTER: { type: "AFTER", value: new Date("2023-04-01T09:24:33.120Z") },
};

const filterByBefore: FiltersObj = {
  BEFORE: { type: "BEFORE", value: new Date("2023-05-23T11:11:21.817Z") },
};

const filterByAsTriggerStatus: FiltersObj = {
  ...filterByAs,
  ...filterByTrigger,
  ...filterByStatus,
};

const filterByTriggerAfterBefore: FiltersObj = {
  ...filterByTrigger,
  ...filterByAfter,
  ...filterByBefore,
};

const filterByAsBefore: FiltersObj = {
  ...filterByAs,
  ...filterByBefore,
};

const queryForAs = "&filter.field=AS&filter.type=CONTAINS&filter.val=Findme";
const queryForStatus =
  "&filter.field=STATUS&filter.type=MATCH&filter.val=failed";
const queryForTrigger =
  "&filter.field=TRIGGER&filter.type=MATCH&filter.val=cloudevent";
const queryForAfter =
  "&filter.field=CREATED&filter.type=AFTER&filter.val=2023-04-01T09:24:33.120Z";
const queryForBefore =
  "&filter.field=CREATED&filter.type=BEFORE&filter.val=2023-05-23T11:11:21.817Z";

describe("getFilterQuery", () => {
  test("it returns a query string for filtering by name", () => {
    expect(getFilterQuery(filterByAs)).toBe(queryForAs);
  });

  test("it returns a query string for filtering by status", () => {
    expect(getFilterQuery(filterByStatus)).toBe(queryForStatus);
  });

  test("it returns a query string for filtering by trigger", () => {
    expect(getFilterQuery(filterByTrigger)).toBe(queryForTrigger);
  });

  test("it returns a query string for filtering by created after", () => {
    expect(getFilterQuery(filterByAfter)).toBe(queryForAfter);
  });

  test("it returns a query string for filtering by created before", () => {
    expect(getFilterQuery(filterByBefore)).toBe(queryForBefore);
  });

  test("it returns a query string for multiple filters: name, trigger, status", () => {
    expect(getFilterQuery(filterByAsTriggerStatus)).toBe(
      queryForAs + queryForTrigger + queryForStatus
    );
  });

  test("it returns a query string for multiple filters: name, before", () => {
    expect(getFilterQuery(filterByAsBefore)).toBe(queryForAs + queryForBefore);
  });

  test("it returns a query string for multiple filters: trigger, after, before", () => {
    expect(getFilterQuery(filterByTriggerAfterBefore)).toBe(
      queryForTrigger + queryForAfter + queryForBefore
    );
  });
});

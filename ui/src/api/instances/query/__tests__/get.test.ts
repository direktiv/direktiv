import { FiltersObj, getFilterQuery } from "../utils";
import { describe, expect, test } from "vitest";

const filterByAs: FiltersObj = {
  AS: { type: "CONTAINS", value: "Findme" },
};

const filterByStatus: FiltersObj = {
  status: { type: "MATCH", value: "failed" },
};

const filterByTrigger: FiltersObj = {
  trigger: { type: "MATCH", value: "cloudevent" },
};

const filterByAfter: FiltersObj = {
  createdAtGt: { operator: "gt", value: new Date("2023-04-01T09:24:33.120Z") },
};

const filterByBefore: FiltersObj = {
  createdAtLt: { operator: "lt", value: new Date("2023-05-23T11:11:21.817Z") },
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

const queryForAs = "&filter[AS]=Findme";
const queryForStatus = "&filter[status]=failed";
const queryForTrigger = "&filter[trigger]=cloudevent";
const queryForAfter = "&filter[createdAt][gt]=2023-04-01T09:24:33.120Z";
const queryForBefore = "&filter[createdAt][lt]=2023-05-23T11:11:21.817Z";

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

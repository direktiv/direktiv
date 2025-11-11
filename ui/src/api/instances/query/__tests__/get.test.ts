import { FiltersObj, getFilterQuery } from "../utils";
import { describe, expect, test } from "vitest";

const filterByPath: FiltersObj = {
  path: { operator: "cn", value: "Findme" },
};

const filterByStatus: FiltersObj = {
  status: { value: "failed" },
};

const filterByTrigger: FiltersObj = {
  trigger: { value: "cloudevent" },
};

const filterByAfter: FiltersObj = {
  createdAtGt: { operator: "gt", value: new Date("2023-04-01T09:24:33.120Z") },
};

const filterByBefore: FiltersObj = {
  createdAtLt: { operator: "lt", value: new Date("2023-05-23T11:11:21.817Z") },
};

const filterByPathTriggerStatus: FiltersObj = {
  ...filterByPath,
  ...filterByTrigger,
  ...filterByStatus,
};

const filterByTriggerAfterBefore: FiltersObj = {
  ...filterByTrigger,
  ...filterByAfter,
  ...filterByBefore,
};

const filterByPathBefore: FiltersObj = {
  ...filterByPath,
  ...filterByBefore,
};

const queryForPath = "&filter[path][cn]=Findme";
const queryForStatus = "&filter[status]=failed";
const queryForTrigger = "&filter[trigger]=cloudevent";
const queryForAfter = "&filter[createdAt][gt]=2023-04-01T09:24:33.120Z";
const queryForBefore = "&filter[createdAt][lt]=2023-05-23T11:11:21.817Z";

describe("getFilterQuery", () => {
  test("it returns a query string for filtering by name", () => {
    expect(getFilterQuery(filterByPath)).toBe(queryForPath);
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
    expect(getFilterQuery(filterByPathTriggerStatus)).toBe(
      queryForPath + queryForTrigger + queryForStatus
    );
  });

  test("it returns a query string for multiple filters: name, before", () => {
    expect(getFilterQuery(filterByPathBefore)).toBe(
      queryForPath + queryForBefore
    );
  });

  test("it returns a query string for multiple filters: trigger, after, before", () => {
    expect(getFilterQuery(filterByTriggerAfterBefore)).toBe(
      queryForTrigger + queryForAfter + queryForBefore
    );
  });
});

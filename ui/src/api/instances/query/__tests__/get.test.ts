import { FiltersObj, getFilterQuery } from "../utils";
import { describe, expect, test } from "vitest";

const filterByPath: FiltersObj = {
  path: { operator: "cn", value: "Findme" },
};

const filterByStatus: FiltersObj = {
  status: { value: "failed" },
};

const filterByInvoker: FiltersObj = {
  invoker: { value: "event" },
};

const filterByAfter: FiltersObj = {
  createdAtGt: { operator: "gt", value: new Date("2023-04-01T09:24:33.120Z") },
};

const filterByBefore: FiltersObj = {
  createdAtLt: { operator: "lt", value: new Date("2023-05-23T11:11:21.817Z") },
};

const filterByPathInvokerStatus: FiltersObj = {
  ...filterByPath,
  ...filterByInvoker,
  ...filterByStatus,
};

const filterByInvokerAfterBefore: FiltersObj = {
  ...filterByInvoker,
  ...filterByAfter,
  ...filterByBefore,
};

const filterByPathBefore: FiltersObj = {
  ...filterByPath,
  ...filterByBefore,
};

const queryForPath = "&filter[path][cn]=Findme";
const queryForStatus = "&filter[status]=failed";
const queryForInvoker = "&filter[invoker]=event";
const queryForAfter = "&filter[createdAt][gt]=2023-04-01T09:24:33.120Z";
const queryForBefore = "&filter[createdAt][lt]=2023-05-23T11:11:21.817Z";

describe("getFilterQuery", () => {
  test("it returns a query string for filtering by name", () => {
    expect(getFilterQuery(filterByPath)).toBe(queryForPath);
  });

  test("it returns a query string for filtering by status", () => {
    expect(getFilterQuery(filterByStatus)).toBe(queryForStatus);
  });

  test("it returns a query string for filtering by invoker", () => {
    expect(getFilterQuery(filterByInvoker)).toBe(queryForInvoker);
  });

  test("it returns a query string for filtering by created after", () => {
    expect(getFilterQuery(filterByAfter)).toBe(queryForAfter);
  });

  test("it returns a query string for filtering by created before", () => {
    expect(getFilterQuery(filterByBefore)).toBe(queryForBefore);
  });

  test("it returns a query string for multiple filters: name, invoker, status", () => {
    expect(getFilterQuery(filterByPathInvokerStatus)).toBe(
      queryForPath + queryForInvoker + queryForStatus
    );
  });

  test("it returns a query string for multiple filters: name, before", () => {
    expect(getFilterQuery(filterByPathBefore)).toBe(
      queryForPath + queryForBefore
    );
  });

  test("it returns a query string for multiple filters: invoker, after, before", () => {
    expect(getFilterQuery(filterByInvokerAfterBefore)).toBe(
      queryForInvoker + queryForAfter + queryForBefore
    );
  });
});

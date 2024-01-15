import { FiltersObj, getFilterQuery } from "../get";
import { describe, expect, test } from "vitest";

const filterByType: FiltersObj = {
  TYPE: { type: "CONTAINS", value: "eventtype" },
};

const filterByText: FiltersObj = {
  TEXT: { type: "CONTAINS", value: "anything" },
};

const filterByAfter: FiltersObj = {
  AFTER: { type: "AFTER", value: new Date("2023-04-01T09:24:33.120Z") },
};

const filterByBefore: FiltersObj = {
  BEFORE: { type: "BEFORE", value: new Date("2023-05-23T11:11:21.817Z") },
};

const filterByTypeBefore: FiltersObj = {
  ...filterByType,
  ...filterByBefore,
};
const filterByTextAfterBefore: FiltersObj = {
  ...filterByText,
  ...filterByAfter,
  ...filterByBefore,
};

const queryForType =
  "&filter.field=TYPE&filter.type=CONTAINS&filter.val=eventtype";
const queryForText =
  "&filter.field=TEXT&filter.type=CONTAINS&filter.val=anything";
const queryForAfter =
  "&filter.field=CREATED&filter.type=AFTER&filter.val=2023-04-01T09:24:33.120Z";
const queryForBefore =
  "&filter.field=CREATED&filter.type=BEFORE&filter.val=2023-05-23T11:11:21.817Z";

describe("getFilterQuery", () => {
  test("it returns a query string for filtering by type", () => {
    expect(getFilterQuery(filterByType)).toBe(queryForType);
  });

  test("it returns a query string for filtering by text", () => {
    expect(getFilterQuery(filterByText)).toBe(queryForText);
  });

  test("it returns a query string for filtering by created after", () => {
    expect(getFilterQuery(filterByAfter)).toBe(queryForAfter);
  });

  test("it returns a query string for filtering by created before", () => {
    expect(getFilterQuery(filterByBefore)).toBe(queryForBefore);
  });

  test("it returns a query string for multiple filters: type, before", () => {
    expect(getFilterQuery(filterByTypeBefore)).toBe(
      queryForType + queryForBefore
    );
  });

  test("it returns a query string for multiple filters: text, after, before", () => {
    expect(getFilterQuery(filterByTextAfterBefore)).toBe(
      queryForText + queryForAfter + queryForBefore
    );
  });
});

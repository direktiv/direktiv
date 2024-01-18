import { FiltersObj, getFilterQuery } from "../get";
import { describe, expect, test } from "vitest";

const filterByState: FiltersObj = {
  stateName: "mystate",
};

const filterByWorkflow: FiltersObj = {
  workflowName: "myWorkflow.yaml",
};

const filterByStateAndWorkflow: FiltersObj = {
  stateName: "mystate",
  workflowName: "myWorkflow.yaml",
};

describe("getFilterQuery", () => {
  test("it returns a query string for filtering by state", () => {
    expect(getFilterQuery(filterByState)).toBe(
      "&filter.field=QUERY&filter.type=MATCH&filter.val=::mystate::"
    );
  });

  test("it returns a query string for filtering by workflow", () => {
    expect(getFilterQuery(filterByWorkflow)).toBe(
      "&filter.field=QUERY&filter.type=MATCH&filter.val=myWorkflow.yaml::::"
    );
  });

  test("it returns a query string for filtering by state and workflow", () => {
    expect(getFilterQuery(filterByStateAndWorkflow)).toBe(
      "&filter.field=QUERY&filter.type=MATCH&filter.val=myWorkflow.yaml::mystate::"
    );
  });
  test("it returns a query string for filtering by state and workflow", () => {
    expect(getFilterQuery()).toBe("");
  });
});

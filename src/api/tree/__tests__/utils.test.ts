import { describe, expect, test } from "vitest";
import {
  forceLeadingSlash,
  removeLeadingSlash,
  removeTrailingSlash,
  sortByName,
  sortFoldersFirst,
} from "../utils";

import { NodeSchemaType } from "../schema";

const itemTemplate: NodeSchemaType = {
  createdAt: "2023-03-13T13:39:05.832664Z",
  updatedAt: "2023-03-13T13:39:06.118436Z",
  name: "demo-workflow",
  path: "/demo-workflow",
  parent: "/",
  type: "directory",
  attributes: [],
  oid: "",
  readOnly: true,
  expandedType: "git",
};

describe("sortFoldersFirst", () => {
  test("will sort all directories to the top, followed by directories and sort them alphabetically", () => {
    const results: NodeSchemaType[] = [
      { ...itemTemplate, name: "workflowB", type: "workflow" },
      { ...itemTemplate, name: "workflowA", type: "workflow" },
      { ...itemTemplate, name: "directoryB", type: "directory" },
      { ...itemTemplate, name: "workflowC", type: "workflow" },
      { ...itemTemplate, name: "directoryA", type: "directory" },
    ];

    const resultSorted = results.sort(sortFoldersFirst);

    expect(resultSorted.map((x) => x.name)).toStrictEqual([
      "directoryA",
      "directoryB",
      "workflowA",
      "workflowB",
      "workflowC",
    ]);
  });
});

describe("sortByName", () =>
  test("will sort array by name", () => {
    const results = [
      { name: "zZ" },
      { name: "zz" },
      { name: "abc" },
      { name: "e" },
      { name: "fa" },
      { name: "f" },
    ];

    const resultSorted = results.sort(sortByName);
    expect(resultSorted.map((x) => x.name)).toStrictEqual([
      "abc",
      "e",
      "f",
      "fa",
      "zz",
      "zZ",
    ]);
  }));

describe("forceLeadingSlash", () => {
  test("path -> /path", () => {
    expect(forceLeadingSlash("path")).toBe("/path");
  });

  test("/path -> /path", () => {
    expect(forceLeadingSlash("/path")).toBe("/path");
  });

  test("empty string -> empty string", () => {
    expect(forceLeadingSlash("")).toBe("/");
  });

  test("undefined -> empty string", () => {
    expect(forceLeadingSlash()).toBe("/");
  });
});

describe("removeLeadingSlash", () => {
  test("/path -> path", () => {
    expect(removeLeadingSlash("/path")).toBe("path");
  });

  test("path -> path", () => {
    expect(removeLeadingSlash("path")).toBe("path");
  });

  test("/ -> empty string", () => {
    expect(removeLeadingSlash("/")).toBe("");
  });

  test("empty string -> empty string", () => {
    expect(removeLeadingSlash("")).toBe("");
  });

  test("undefined -> empty string", () => {
    expect(removeLeadingSlash()).toBe("");
  });
});

describe("removeTrailingSlash", () => {
  test("path/ -> path", () => {
    expect(removeTrailingSlash("path/")).toBe("path");
  });

  test("path -> path", () => {
    expect(removeTrailingSlash("path")).toBe("path");
  });

  test("/ -> empty string", () => {
    expect(removeTrailingSlash("/")).toBe("");
  });

  test("empty string -> empty string", () => {
    expect(removeTrailingSlash("")).toBe("");
  });

  test("undefined -> empty string", () => {
    expect(removeTrailingSlash()).toBe("");
  });
});

import { describe, expect, test } from "vitest";
import { forceSlashIfPath, sortFoldersFirst } from "../utils";

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
  test("will sort all directories to the top, followed by directories and sort them alphabetically ", () => {
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

describe("forceSlashIfPath", () => {
  test("path -> /path", () => {
    expect(forceSlashIfPath("path")).toBe("/path");
  });

  test("/path -> /path", () => {
    expect(forceSlashIfPath("/path")).toBe("/path");
  });

  test("undefined -> empty string", () => {
    expect(forceSlashIfPath()).toBe("");
  });
});

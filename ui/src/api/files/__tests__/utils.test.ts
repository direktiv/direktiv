import { describe, expect, test } from "vitest";
import {
  forceLeadingSlash,
  getFilenameFromPath,
  getParentFromPath,
  removeLeadingSlash,
  removeTrailingSlash,
  sortByName,
  sortFoldersFirst,
} from "../utils";

import { BaseFileSchemaType } from "../schema";

const itemTemplate: BaseFileSchemaType = {
  createdAt: "2023-03-13T13:39:05.832664Z",
  updatedAt: "2023-03-13T13:39:06.118436Z",
  path: "/demo-workflow",
  type: "directory",
};

describe("getFilenameFromPath()", () => {
  test("file at root level", () => {
    expect(getFilenameFromPath("/action.yaml")).toBe("action.yaml");
  });
  test("file in directory", () => {
    expect(getFilenameFromPath("/foobar/nested.yaml")).toBe("nested.yaml");
  });
  test("directory", () => {
    expect(getFilenameFromPath("/examples")).toBe("examples");
  });
  test("nested directory", () => {
    expect(getFilenameFromPath("/examples/nested")).toBe("nested");
  });
});

describe("getParentFromPath()", () => {
  test("file at root level", () => {
    expect(getParentFromPath("/action.yaml")).toBe("/");
  });
  test("file in directory", () => {
    expect(getParentFromPath("/foobar/nested.yaml")).toBe("/foobar");
  });
  test("directory at root level", () => {
    expect(getParentFromPath("/examples")).toBe("/");
  });
  test("nested directory", () => {
    expect(getParentFromPath("/examples/nested")).toBe("/examples");
  });
  test("empty string", () => {
    expect(() => getParentFromPath("")).toThrowError(
      "Cannot infer parent from empty string"
    );
  });
  test("root", () => {
    expect(() => getParentFromPath("/")).toThrowError(
      "Cannot infer parent from '/'"
    );
  });
});

describe("sortFoldersFirst", () => {
  test("will sort all directories to the top, followed by directories and sort them alphabetically", () => {
    const results: BaseFileSchemaType[] = [
      { ...itemTemplate, path: "/b-service.yaml", type: "service" },
      { ...itemTemplate, path: "/c-workflow.yaml", type: "workflow" },
      { ...itemTemplate, path: "/b-directory", type: "directory" },
      { ...itemTemplate, path: "/a-workflow.yaml", type: "workflow" },
      { ...itemTemplate, path: "/a-directory", type: "directory" },
    ];

    const resultSorted = results.sort(sortFoldersFirst);

    expect(resultSorted.map((x) => x.path)).toStrictEqual([
      "/a-directory",
      "/b-directory",
      "/a-workflow.yaml",
      "/b-service.yaml",
      "/c-workflow.yaml",
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

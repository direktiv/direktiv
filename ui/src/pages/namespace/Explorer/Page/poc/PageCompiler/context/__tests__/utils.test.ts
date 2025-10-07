import { ParentBlockType, SimpleBlockType } from "../../../schema/blocks";
import { describe, expect, test } from "vitest";
import {
  findAncestor,
  findBlock,
  incrementPath,
  isFirstChildPath,
  isPage,
  isParentBlock,
  pathIsDescendant,
  pathsEqual,
  reindexTargetPath,
} from "../utils";

import { BlockPathType } from "../../Block";
import { ColumnsType } from "../../../schema/blocks/columns";
import { DirektivPagesType } from "../../../schema";
import complex from "../../../schema/__tests__/examples/complex";
import simple from "../../../schema/__tests__/examples/simple";

const parentBlock = {
  type: "columns",
  blocks: [
    {
      type: "column",
      blocks: [{ type: "text", content: "some text goes here" }],
    },
    {
      type: "column",
      blocks: [{ type: "text", content: "some text goes here" }],
    },
  ],
} satisfies ParentBlockType;

const simpleBlock = {
  type: "headline",
  label: "Lorem ipsum",
  level: "h1",
} satisfies SimpleBlockType;

describe("isParentBlock", () => {
  test("it should return false for a page", () => {
    const falseParentBlock = simple as unknown;
    const result = isParentBlock(falseParentBlock as ParentBlockType);
    expect(result).toEqual(false);
  });

  test("it should return false for a simple block", () => {
    const result = isParentBlock(simpleBlock);
    expect(result).toEqual(false);
  });

  test("it should return true for a parent block", () => {
    const result = isParentBlock(parentBlock);
    expect(result).toEqual(true);
  });
});

describe("isFirstChildPath", () => {
  test("it should return true for the first element on the page", () => {
    const rootBlockPath = [0];

    const result = isFirstChildPath(rootBlockPath);
    expect(result).toEqual(true);
  });

  test("it should return true for the first child of the first element", () => {
    const rootChildBlockPath = [0, 0, 0];

    const result = isFirstChildPath(rootChildBlockPath);
    expect(result).toEqual(true);
  });

  test("it should return true for the first element in a column", () => {
    const firstColumnBlockPath = [2, 0, 0];

    const result = isFirstChildPath(firstColumnBlockPath);
    expect(result).toEqual(true);
  });

  test("it should return false for the second element in a column", () => {
    const secondColumnBlockPath = [2, 0, 1];

    const result = isFirstChildPath(secondColumnBlockPath);
    expect(result).toEqual(false);
  });
});

describe("isPage", () => {
  test("it should return true for a page", () => {
    const result = isPage(simple);
    expect(result).toEqual(true);
  });

  test("it should return false for a simple block", () => {
    const result = isPage(simpleBlock);
    expect(result).toEqual(false);
  });

  test("it should return false for a parent block", () => {
    const result = isPage(parentBlock);
    expect(result).toEqual(false);
  });
});

describe("findBlock", () => {
  test("it returns the block at the specified path", () => {
    const result = findBlock(complex, [2, 1, 0]);
    expect(result).toEqual({
      type: "text",
      content: "Column 2 text",
    });
  });

  test("it returns the block including nested blocks (1)", () => {
    const result = findBlock(complex, [3]);
    expect(result).toEqual({
      type: "query-provider",
      queries: [
        {
          id: "fetching-resources",
          url: "/api/get/resources",
          queryParams: [
            {
              key: "query",
              value: "my-search-query",
            },
          ],
        },
      ],
      blocks: [
        {
          type: "dialog",
          trigger: {
            type: "button",
            label: "open dialog",
          },
          blocks: [
            {
              type: "form",
              trigger: {
                type: "button",
                label: "delete",
              },
              mutation: {
                url: "/api/delete/",
                method: "DELETE",
              },
              blocks: [],
            },
          ],
        },
        {
          type: "text",
          content: "simple text",
        },
      ],
    });
  });

  test("it returns the block including nested blocks (2)", () => {
    const result = findBlock(complex, [3, 0]);
    expect(result).toEqual({
      type: "dialog",
      trigger: {
        type: "button",
        label: "open dialog",
      },
      blocks: [
        {
          type: "form",
          trigger: {
            type: "button",
            label: "delete",
          },
          mutation: {
            url: "/api/delete/",
            method: "DELETE",
          },
          blocks: [],
        },
      ],
    });
  });

  test("it returns the whole input object if the path is empty", () => {
    const result = findBlock(complex, []);
    expect(result).toEqual(complex);
  });

  test("it throws an error if an index in the path is not found", () => {
    expect(() => findBlock(complex, [5, 3, 0])).toThrow("index 5 not found");
  });
});

describe("reindexTargetPath", () => {
  test("decrements path segment when target > origin at reindex level", () => {
    const origin = [1, 0, 3];
    const target = [1, 0, 7];
    const result = reindexTargetPath(origin, target);
    expect(result).toEqual([1, 0, 6]);
  });

  test("returns unchanged copy when target < origin at reindex level", () => {
    const origin = [1, 0, 6];
    const target = [1, 0, 2];
    const result = reindexTargetPath(origin, target);
    expect(result).toEqual([1, 0, 2]);
    expect(result).not.toBe(target);
  });

  test("returns unchanged copy when moving between deeply nested levels in different branches", () => {
    const origin = [1, 3, 6, 1];
    const target = [1, 4, 6, 3];
    const result = reindexTargetPath(origin, target);
    expect(result).toEqual([1, 4, 6, 3]);
    expect(result).not.toBe(target);
  });

  test("handles moving to shallower path where target = origin at reindex level", () => {
    const origin = [1, 0, 6, 3];
    const target = [1, 0, 6];
    const result = reindexTargetPath(origin, target);
    expect(result).toEqual([1, 0, 6]);
    expect(result).not.toBe(target);
  });

  test("handles moving to shallower path where target > origin at reindex level", () => {
    const origin = [2, 4, 5, 6, 7];
    const target = [2, 4, 6];
    const result = reindexTargetPath(origin, target);
    expect(result).toEqual([2, 4, 6]);
  });

  test("handles moving to shallower path where target < origin at reindex level", () => {
    const origin = [2, 4, 5, 6, 7];
    const target = [2, 4, 3];
    const result = reindexTargetPath(origin, target);
    expect(result).toEqual([2, 4, 3]);
  });

  test("handles moving to deeper path where target > origin at reindex level", () => {
    const origin = [1, 3, 4];
    const target = [1, 3, 6, 1];
    const result = reindexTargetPath(origin, target);
    expect(result).toEqual([1, 3, 5, 1]);
  });

  test("handles moving on root level where target > origin", () => {
    const origin = [4];
    const target = [7];
    const result = reindexTargetPath(origin, target);
    expect(result).toEqual([6]);
  });

  test("handles moving on root level where origin > target", () => {
    const origin = [5];
    const target = [2];
    const result = reindexTargetPath(origin, target);
    expect(result).toEqual([2]);
  });

  test("throws when origin and target path are exactly equal", () => {
    const origin = [3, 2, 1];
    const target = [3, 2, 1];
    expect(() => reindexTargetPath(origin, target)).toThrow(
      "Origin and target paths must not be equal"
    );
  });

  test("throws when origin path is empty", () => {
    const origin: BlockPathType = [];
    const target = [3, 2, 1];
    expect(() => reindexTargetPath(origin, target)).toThrow(
      "Paths must not be empty"
    );
  });

  test("throws when target path is empty", () => {
    const origin = [9];
    const target: BlockPathType = [];
    expect(() => reindexTargetPath(origin, target)).toThrow(
      "Paths must not be empty"
    );
  });
});

describe("incrementPath", () => {
  test("should increment the last index of a non-empty path", () => {
    const input = [0, 2, 2];
    const expected = [0, 2, 3];
    expect(incrementPath(input)).toEqual(expected);
  });

  test("should handle a path with a single element", () => {
    const input = [0];
    const expected = [1];
    expect(incrementPath(input)).toEqual(expected);
  });

  test("should handle a path with last index 0", () => {
    const input = [2, 0];
    const expected = [2, 1];
    expect(incrementPath(input)).toEqual(expected);
  });

  test("should return an empty array unchanged", () => {
    const input: number[] = [];
    const expected: number[] = [];
    expect(incrementPath(input)).toEqual(expected);
  });
});

describe("pathIsDescendant", () => {
  test("returns true when descendant starts with ancestor", () => {
    expect(pathIsDescendant([0, 4, 3, 1], [0, 4, 3])).toBe(true);
    expect(pathIsDescendant([1, 2, 3], [1])).toBe(true);
    expect(pathIsDescendant([5, 6, 7, 8], [5, 6])).toBe(true);
  });

  test("returns true when ancestor is []", () => {
    expect(pathIsDescendant([0], [])).toBe(true);
  });

  test("returns false when descendant and ancestor are exactly equal", () => {
    expect(pathIsDescendant([1, 2, 3], [1, 2, 3])).toBe(false);
    expect(pathIsDescendant([], [])).toBe(false);
  });

  test("returns false when ancestor is longer than descendant", () => {
    expect(pathIsDescendant([1, 2], [1, 2, 3])).toBe(false);
  });

  test("returns false when descendant does not start with ancestor", () => {
    expect(pathIsDescendant([0, 4, 3, 1], [4, 3])).toBe(false);
    expect(pathIsDescendant([1, 2, 3], [2])).toBe(false);
  });
});

describe("findAncestor", () => {
  test("it does not consider the last index in the path", () => {
    const result = findAncestor({
      page: complex,
      path: [2, 0, 0],
      match: (block) => block.type && block.type === "text",
    });
    expect(result).toEqual(null);
  });

  test("it returns an object if an ancester matches fn", () => {
    const target = complex.blocks[2];

    const result = findAncestor({
      page: complex,
      path: [2, 0, 0],
      match: (block) => block.type && block.type === "columns",
    });
    expect(result).toEqual({
      block: target,
      path: [2],
    });
  });

  test("it returns null if no element in the branch upwards matches the fn", () => {
    const result = findAncestor({
      page: complex,
      path: [2, 0],
      match: (block) => block.type && block.type === "text",
    });
    expect(result).toEqual(null);
  });

  test("it returns an object if depth is 1 and the first ancestor matches fn", () => {
    const target = (complex.blocks[2] as ColumnsType).blocks[0];

    const result = findAncestor({
      page: complex,
      path: [2, 0, 0],
      match: (block) => block.type && block.type === "column",
      depth: 1,
    });
    expect(result).toEqual({ block: target, path: [2, 0] });
  });

  test("it returns null if depth is 1 and the first ancestor does not match fn", () => {
    const result = findAncestor({
      page: complex,
      path: [2, 0, 0],
      match: (block) => block.type && block.type === "text",
      depth: 1,
    });
    expect(result).toEqual(null);
  });

  test("it returns true if elements within depth 2 evaluate as true", () => {
    const target = complex.blocks[3];
    const result = findAncestor({
      page: complex,
      path: [3, 0, 0],
      match: (block) => block.type && block.type === "query-provider",
      depth: 2,
    });
    expect(result).toEqual({
      block: target,
      path: [3],
    });
  });

  test("it returns null if elements within depth 2 evaluate as false", () => {
    const result = findAncestor({
      page: complex,
      path: [3, 0, 0],
      match: (block) => block.type && block.type === "columns",
      depth: 2,
    });
    expect(result).toEqual(null);
  });

  test("it evaluates correctly if path is [0]", () => {
    const page = {
      direktiv_api: "page/v1",
      type: "page",
      blocks: [
        {
          type: "card",
          blocks: [],
        },
      ],
    } satisfies DirektivPagesType;
    const result = findAncestor({
      page,
      path: [0],
      match: (block) => block.type && block.type === "page",
    });
    expect(result).toEqual({
      block: page,
      path: [],
    });
  });

  test("it evaluates correctly (null) if path is [0]", () => {
    const page = {
      direktiv_api: "page/v1",
      type: "page",
      blocks: [
        {
          type: "card",
          blocks: [],
        },
      ],
    } satisfies DirektivPagesType;
    const result = findAncestor({
      page,
      path: [0],
      match: (block) => block.type && block.type === "card",
    });
    expect(result).toEqual(null);
  });

  test("it throws an error if depth is 0", () => {
    expect(() =>
      findAncestor({
        page: complex,
        path: [2, 0, 0],
        match: (block) => block.type && block.type === "column",
        depth: 0,
      })
    ).toThrow("depth must be undefined or >= 1");
  });

  test("it throws an error if depth is negative", () => {
    expect(() =>
      findAncestor({
        page: complex,
        path: [2, 0, 0],
        match: (block) => block.type && block.type === "column",
        depth: -1,
      })
    ).toThrow("depth must be undefined or >= 1");
  });

  test("it evaluates as null when path is [], as there are no parents", () => {
    const page = {
      direktiv_api: "page/v1",
      type: "page",
      blocks: [
        {
          type: "card",
          blocks: [],
        },
      ],
    } satisfies DirektivPagesType;
    const result = findAncestor({
      page,
      path: [],
      match: (block) => block.type && block.type === "page",
    });
    expect(result).toEqual(null);
  });
});

describe("pathsEqual", () => {
  test("it returns true for two matching arrays of numbers", () => {
    const result = pathsEqual([0, 3, 2, 5], [0, 3, 2, 5]);
    expect(result).toEqual(true);
  });

  test("it returns false for arrays of different lengths", () => {
    const result = pathsEqual([0, 3, 2, 5], [0, 3, 2, 5, 1]);
    expect(result).toEqual(false);
  });

  test("it returns false for arrays containing different numbers", () => {
    const result = pathsEqual([0, 3, 2, 5], [0, 2, 3, 5]);
    expect(result).toEqual(false);
  });

  test("it accepts null values (and returns false if they do not match)", () => {
    const result = pathsEqual(null, [0, 2, 3, 5]);
    expect(result).toEqual(false);
  });

  test("it accepts null values (and returns false if they do not match)", () => {
    const result = pathsEqual([0, 2, 3, 5], null);
    expect(result).toEqual(false);
  });

  test("it accepts null values (and returns true if they match)", () => {
    const result = pathsEqual(null, null);
    expect(result).toEqual(true);
  });

  test("it accepts empty arrays (and returns true if they match)", () => {
    const result = pathsEqual([], []);
    expect(result).toEqual(true);
  });
});

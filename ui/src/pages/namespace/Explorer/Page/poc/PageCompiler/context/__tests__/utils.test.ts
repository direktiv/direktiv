import { ColumnType, ColumnsType } from "../../../schema/blocks/columns";
import { ParentBlocksType, SimpleBlocksType } from "../../../schema/blocks";
import {
  addBlockToPage,
  deleteBlockFromPage,
  findAncestor,
  findBlock,
  isPage,
  isParentBlock,
  pathsEqual,
  updateBlockInPage,
} from "../utils";

import { describe, expect, test } from "vitest";
import { DirektivPagesType } from "../../../schema";
import { HeadlineType } from "../../../schema/blocks/headline";
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
} satisfies ParentBlocksType;

const simpleBlock = {
  type: "headline",
  label: "Lorem ipsum",
  level: "h1",
} satisfies SimpleBlocksType;

describe("isParentBlock", () => {
  test("it should return false for a page", () => {
    const falseParentBlock = simple as unknown;
    const result = isParentBlock(falseParentBlock as ParentBlocksType);
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

  test("it returns the whole input object if the path is empty", () => {
    const result = findBlock(complex, []);
    expect(result).toEqual(complex);
  });

  test("it throws an error if an index in the path is not found", () => {
    expect(() => findBlock(complex, [5, 3, 0])).toThrow("index 5 not found");
  });
});

describe("updateBlockInPage", () => {
  test("it updates the block at the specified path", () => {
    const result = updateBlockInPage(complex, [2, 1, 1], {
      type: "text",
      content: "I am now a text block",
    });

    const targetBlock = (
      (result.blocks[2] as ColumnsType).blocks[1] as ColumnType
    ).blocks[1];

    expect(targetBlock).toEqual({
      type: "text",
      content: "I am now a text block",
    });
  });

  test("it throws an error if an empty array is given as index", () => {
    expect(() =>
      updateBlockInPage(simple, [], {
        type: "text",
        content: "I am now a text block",
      })
    ).toThrow("Invalid path, could not extract index for target block");
  });

  test("it throws an error if the target block index is out of bounds", () => {
    expect(() =>
      updateBlockInPage(complex, [2, 1, 9], {
        type: "text",
        content: "I am now a text block",
      })
    ).toThrow("Index for updating block out of bounds");
  });

  test("it throws an error if the page is not valid", () => {
    const fakePage = {
      type: "this makes no sense",
      lorem: "ipsum",
    } as unknown;

    expect(() =>
      updateBlockInPage(fakePage as DirektivPagesType, [2, 1, 1], {
        type: "text",
        content: "I am now a text block",
      })
    ).toThrow("index 2 not found");
  });
});

describe("addBlockToPage", () => {
  const headline: HeadlineType = {
    type: "headline",
    label: "New headline",
    level: "h2",
  };

  test("it adds a block at the specified index", () => {
    const result = addBlockToPage(simple, [2, 1, 0], headline);

    expect((result.blocks[2] as ColumnsType).blocks[1]).toEqual({
      type: "column",
      blocks: [
        {
          type: "headline",
          label: "New headline",
          level: "h2",
        },
        {
          type: "text",
          content: "second column text",
        },
      ],
    });
  });

  test("it adds a block after the specified index", () => {
    const result = addBlockToPage(simple, [2, 1, 0], headline, true);

    expect((result.blocks[2] as ColumnsType).blocks[1]).toEqual({
      type: "column",
      blocks: [
        {
          type: "text",
          content: "second column text",
        },
        {
          type: "headline",
          label: "New headline",
          level: "h2",
        },
      ],
    });
  });

  test("it throws an error if the specified index is empty", () => {
    expect(() => addBlockToPage(simple, [], headline)).toThrow(
      "Invalid path, could not extract index for new block"
    );
  });
});

describe("deleteBlockFromPage", () => {
  test("it deletes a block from the page", () => {
    const result = deleteBlockFromPage(simple, [2, 0, 0]);
    expect(result.blocks[2] as ColumnsType).toEqual({
      type: "columns",
      blocks: [
        {
          type: "column",
          blocks: [],
        },
        {
          type: "column",
          blocks: [{ type: "text", content: "second column text" }],
        },
      ],
    });
  });

  test("it throws an error if an empty array is given as index", () => {
    expect(() => deleteBlockFromPage(simple, [])).toThrow(
      "Invalid path, could not extract index for target block"
    );
  });
});

describe("findAncestor", () => {
  test("it does not consider the last index in the path", () => {
    const result = findAncestor({
      page: complex,
      path: [2, 0, 0],
      match: (block) => block.type && block.type === "text",
    });
    expect(result).toEqual(false);
  });

  test("it returns true if an ancester matches fn", () => {
    const result = findAncestor({
      page: complex,
      path: [2, 0, 0],
      match: (block) => block.type && block.type === "columns",
    });
    expect(result).toEqual(true);
  });

  test("it returns false if no element in the branch upwards matches the fn", () => {
    const result = findAncestor({
      page: complex,
      path: [2, 0],
      match: (block) => block.type && block.type === "text",
    });
    expect(result).toEqual(false);
  });

  test("it returns true if depth is 1 and the first ancestor matches fn", () => {
    const result = findAncestor({
      page: complex,
      path: [2, 0, 0],
      match: (block) => block.type && block.type === "column",
      depth: 1,
    });
    expect(result).toEqual(true);
  });

  test("it returns false if depth is 1 and the first ancestor does not match fn", () => {
    const result = findAncestor({
      page: complex,
      path: [2, 0, 0],
      match: (block) => block.type && block.type === "text",
      depth: 1,
    });
    expect(result).toEqual(false);
  });

  test("it returns true if elements within depth 2 evaluate as true", () => {
    const result = findAncestor({
      page: complex,
      path: [3, 0, 0],
      match: (block) => block.type && block.type === "dialog",
      depth: 2,
    });
    expect(result).toEqual(true);
  });

  test("it returns true if elements within depth 3 evaluate as true", () => {
    const result = findAncestor({
      page: complex,
      path: [3, 0, 0],
      match: (block) => block.type && block.type === "query-provider",
      depth: 3,
    });
    expect(result).toEqual(true);
  });

  test("it returns false if elements within depth 3 evaluate as false", () => {
    const result = findAncestor({
      page: complex,
      path: [3, 0, 0],
      match: (block) => block.type && block.type === "columns",
      depth: 3,
    });
    expect(result).toEqual(false);
  });

  test("it evaluates correctly (true) if path is [0]", () => {
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
    expect(result).toEqual(true);
  });

  test("it evaluates correctly (false) if path is [0]", () => {
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
    expect(result).toEqual(false);
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

  test("it evaluates as false when path is [], as there are no parents", () => {
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
    expect(result).toEqual(false);
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

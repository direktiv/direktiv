import { ParentBlocksType, SimpleBlocksType } from "../../../schema/blocks";
import { describe, expect, test } from "vitest";

import { isPage, isParentBlock } from "../utils";
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

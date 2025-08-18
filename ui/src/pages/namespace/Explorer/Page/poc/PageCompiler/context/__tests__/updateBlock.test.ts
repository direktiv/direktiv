import { ColumnType, ColumnsType } from "../../../schema/blocks/columns";
import { describe, expect, test } from "vitest";

import { DirektivPagesType } from "../../../schema";
import complex from "../../../schema/__tests__/examples/complex";
import simple from "../../../schema/__tests__/examples/simple";
import { updateBlockInPage } from "../utils/updatePage";

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

import { describe, expect, test } from "vitest";

import { ColumnsType } from "../../../schema/blocks/columns";
import { deleteBlockFromPage } from "../utils/updatePage";
import simple from "../../../schema/__tests__/examples/simple";

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

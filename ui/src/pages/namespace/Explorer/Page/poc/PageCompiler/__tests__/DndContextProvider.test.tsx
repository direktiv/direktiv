import { describe, expect, test } from "vitest";

import { BlockPathType } from "../Block";
import { DirektivPagesType } from "../../schema";
import { HeadlineType } from "../../schema/blocks/headline";
import { addBlockToPage } from "../context/utils";

describe("addBlockToPage", () => {
  const newBlock: HeadlineType = {
    type: "headline",
    level: "h1",
    label: "NEW HEADLINE",
  };

  const path: BlockPathType = [0];

  const pathLevelBelow: BlockPathType = [0, 0];
  const pathLevelBelowBefore: BlockPathType = [0, 0];
  const pathLevelBelowAfter: BlockPathType = [0, 1];

  test("inserts block on the same level", () => {
    const page: DirektivPagesType = {
      direktiv_api: "page/v1",
      type: "page",
      blocks: [{ type: "text", content: "Original Block" }],
    };

    const updatedPage = addBlockToPage(page, path, newBlock);

    expect(updatedPage.blocks.length).toBe(2);
    expect(updatedPage.blocks[0]).toEqual(newBlock);
  });

  test("inserts block in another block (level below)", () => {
    const page: DirektivPagesType = {
      direktiv_api: "page/v1",
      type: "page",
      blocks: [{ type: "card", blocks: [] }],
    };

    const updatedPage = addBlockToPage(page, pathLevelBelow, newBlock);

    expect(updatedPage).toEqual({
      direktiv_api: "page/v1",
      type: "page",
      blocks: [
        {
          type: "card",
          blocks: [
            {
              type: "headline",
              level: "h1",
              label: "NEW HEADLINE",
            },
          ],
        },
      ],
    });
  });

  test("inserts block in another block (level below) - insert before ", () => {
    const page: DirektivPagesType = {
      direktiv_api: "page/v1",
      type: "page",
      blocks: [
        { type: "card", blocks: [{ type: "text", content: "Original Block" }] },
      ],
    };

    const updatedPage = addBlockToPage(page, pathLevelBelowBefore, newBlock);

    expect(updatedPage).toEqual({
      direktiv_api: "page/v1",
      type: "page",
      blocks: [
        {
          type: "card",
          blocks: [
            {
              type: "headline",
              level: "h1",
              label: "NEW HEADLINE",
            },
            { type: "text", content: "Original Block" },
          ],
        },
      ],
    });
  });

  test("inserts block in another block (level below) - insert after ", () => {
    const page: DirektivPagesType = {
      direktiv_api: "page/v1",
      type: "page",
      blocks: [
        { type: "card", blocks: [{ type: "text", content: "Original Block" }] },
      ],
    };

    const updatedPage = addBlockToPage(page, pathLevelBelowAfter, newBlock);

    expect(updatedPage).toEqual({
      direktiv_api: "page/v1",
      type: "page",
      blocks: [
        {
          type: "card",
          blocks: [
            { type: "text", content: "Original Block" },
            {
              type: "headline",
              level: "h1",
              label: "NEW HEADLINE",
            },
          ],
        },
      ],
    });
  });
});

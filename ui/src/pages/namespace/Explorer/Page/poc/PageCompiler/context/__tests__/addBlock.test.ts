import { describe, expect, test } from "vitest";

import { HeadlineType } from "../../../schema/blocks/headline";
import { addBlockToPage } from "../../../BlockEditor/utils/updatePage";
import simple from "../../../schema/__tests__/examples/simple";

describe("addBlockToPage", () => {
  const headline: HeadlineType = {
    type: "headline",
    label: "New headline",
    level: "h2",
  };

  test("it adds a block at the specified index", () => {
    const updatedPage = addBlockToPage(simple, [2, 1, 0], headline);

    expect(updatedPage).toEqual({
      direktiv_api: "page/v1",
      type: "page",
      blocks: [
        {
          type: "headline",
          level: "h1",
          label: "Welcome to Direktiv",
        },
        {
          type: "text",
          content:
            "This is a block that contains longer text. You might write some Terms and Conditions here or something similar",
        },
        {
          type: "columns",
          blocks: [
            {
              type: "column",
              blocks: [{ type: "text", content: "first column text" }],
            },
            {
              type: "column",
              blocks: [
                {
                  type: "headline",
                  label: "New headline",
                  level: "h2",
                },
                { type: "text", content: "second column text" },
              ],
            },
          ],
        },
      ],
    });
  });

  test("it adds a block after the specified index", () => {
    const updatedPage = addBlockToPage(simple, [2, 1, 0], headline, true);

    expect(updatedPage).toEqual({
      direktiv_api: "page/v1",
      type: "page",
      blocks: [
        {
          type: "headline",
          level: "h1",
          label: "Welcome to Direktiv",
        },
        {
          type: "text",
          content:
            "This is a block that contains longer text. You might write some Terms and Conditions here or something similar",
        },
        {
          type: "columns",
          blocks: [
            {
              type: "column",
              blocks: [{ type: "text", content: "first column text" }],
            },
            {
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
            },
          ],
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

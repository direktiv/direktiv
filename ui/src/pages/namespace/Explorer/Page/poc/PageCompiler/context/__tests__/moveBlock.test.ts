import { describe, expect, test } from "vitest";

import { BlockType } from "../../../schema/blocks";
import complex from "../../../schema/__tests__/examples/complex";
import { findBlock } from "../utils";
import { moveBlockWithinPage } from "../utils/updatePage";

describe("moveBlockWithinPage", () => {
  test("Places block after origin index at root level", () => {
    const block = findBlock(complex, [0]) as BlockType;
    const result = moveBlockWithinPage(complex, [0], [2], block);

    expect(result).toEqual({
      direktiv_api: "page/v1",
      type: "page",
      blocks: [
        {
          type: "text",
          content:
            "This is a block that contains longer text. You might write some Terms and Conditions here or something similar",
        },
        {
          type: "headline",
          level: "h1",
          label: "Welcome to Direktiv",
        },
        {
          type: "columns",
          blocks: [
            {
              type: "column",
              blocks: [{ type: "text", content: "Column 1 text" }],
            },
            {
              type: "column",
              blocks: [
                { type: "text", content: "Column 2 text" },
                { type: "button", label: "Edit me" },
              ],
            },
          ],
        },
        {
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
                    id: "my-delete",
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
        },
      ],
    });
  });

  test("Places block before origin index at root level", () => {
    const block = findBlock(complex, [3]) as BlockType;
    const result = moveBlockWithinPage(complex, [3], [1], block);

    expect(result).toEqual({
      direktiv_api: "page/v1",
      type: "page",
      blocks: [
        {
          type: "headline",
          level: "h1",
          label: "Welcome to Direktiv",
        },
        {
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
                    id: "my-delete",
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
              blocks: [{ type: "text", content: "Column 1 text" }],
            },
            {
              type: "column",
              blocks: [
                { type: "text", content: "Column 2 text" },
                { type: "button", label: "Edit me" },
              ],
            },
          ],
        },
      ],
    });
  });

  test("Moves from nested to root level, before index of own ancestor", () => {
    const block = findBlock(complex, [2, 1, 1]) as BlockType;
    const result = moveBlockWithinPage(complex, [2, 1, 1], [1], block);

    expect(result).toEqual({
      direktiv_api: "page/v1",
      type: "page",
      blocks: [
        {
          type: "headline",
          level: "h1",
          label: "Welcome to Direktiv",
        },
        {
          type: "button",
          label: "Edit me",
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
              blocks: [{ type: "text", content: "Column 1 text" }],
            },
            {
              type: "column",
              blocks: [{ type: "text", content: "Column 2 text" }],
            },
          ],
        },
        {
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
                    id: "my-delete",
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
        },
      ],
    });
  });

  test("Moves from deeper to shallower nested level, at index of own ancestor", () => {
    const block = findBlock(complex, [2, 1, 1]) as BlockType;
    const result = moveBlockWithinPage(complex, [2, 1, 1], [2], block);

    expect(result).toEqual({
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
          type: "button",
          label: "Edit me",
        },
        {
          type: "columns",
          blocks: [
            {
              type: "column",
              blocks: [{ type: "text", content: "Column 1 text" }],
            },

            {
              type: "column",
              blocks: [{ type: "text", content: "Column 2 text" }],
            },
          ],
        },
        {
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
                    id: "my-delete",
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
        },
      ],
    });
  });

  test("Moves from deeper to shallower nested level, after index of own ancestor", () => {
    const block = findBlock(complex, [2, 0, 0]) as BlockType;
    const result = moveBlockWithinPage(complex, [2, 0, 0], [3], block);

    expect(result).toEqual({
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
              blocks: [],
            },
            {
              type: "column",
              blocks: [
                { type: "text", content: "Column 2 text" },
                { type: "button", label: "Edit me" },
              ],
            },
          ],
        },
        {
          type: "text",
          content: "Column 1 text",
        },
        {
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
                    id: "my-delete",
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
        },
      ],
    });
  });

  test("Moves from root to nested level, to parent before its origin position", () => {
    const block = findBlock(complex, [3]) as BlockType;
    const result = moveBlockWithinPage(complex, [3], [2, 1, 1], block);

    expect(result).toEqual({
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
              blocks: [{ type: "text", content: "Column 1 text" }],
            },
            {
              type: "column",
              blocks: [
                { type: "text", content: "Column 2 text" },
                {
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
                            id: "my-delete",
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
                },
                { type: "button", label: "Edit me" },
              ],
            },
          ],
        },
      ],
    });
  });

  test("Moves from root to nested level, to parent after its origin position", () => {
    const block = findBlock(complex, [0]) as BlockType;
    const result = moveBlockWithinPage(complex, [0], [2, 0, 0], block);

    expect(result).toEqual({
      direktiv_api: "page/v1",
      type: "page",
      blocks: [
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
              blocks: [
                {
                  type: "headline",
                  level: "h1",
                  label: "Welcome to Direktiv",
                },
                {
                  type: "text",
                  content: "Column 1 text",
                },
              ],
            },
            {
              type: "column",
              blocks: [
                { type: "text", content: "Column 2 text" },
                { type: "button", label: "Edit me" },
              ],
            },
          ],
        },
        {
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
                    id: "my-delete",
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
        },
      ],
    });
  });
});

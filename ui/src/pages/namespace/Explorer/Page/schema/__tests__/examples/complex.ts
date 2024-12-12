import { DirektivPagesType } from "../..";

export default {
  direktiv_api: "pages/v1",
  path: "/som/path",
  blocks: [
    {
      type: "headline",
      label: "Welcome to Direktiv",
      description: "This is a headline block inside a Direktiv page",
    },
    {
      type: "text",
      label:
        "This is a block that contains longer text. You might write some Terms and Conditions here or something similar",
    },
    {
      type: "two-columns",
      leftBlocks: [
        {
          type: "text",
          label: "Some text goes here",
        },
      ],
      rightBlocks: [
        {
          type: "text",
          label: "Some text goes here",
        },
      ],
    },
    {
      type: "query-provider",
      query: {
        id: "fetching-resources",
        endpoint: "/api/get/resources",
      },
      blocks: [
        {
          type: "modal",
          trigger: {
            type: "button",
            label: "open modal",
          },
          blocks: [
            {
              type: "form",
              trigger: {
                type: "button",
                label: "delte",
              },
              mutation: {
                id: "my-delete",
                endpoint: "/api/delete/",
                method: "DELETE",
              },
              blocks: [],
            },
          ],
        },
      ],
    },
  ],
} satisfies DirektivPagesType;

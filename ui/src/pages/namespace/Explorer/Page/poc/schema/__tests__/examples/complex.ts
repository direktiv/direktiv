import { DirektivPagesType } from "../..";

export default {
  direktiv_api: "page/v1",
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
          blocks: [
            {
              type: "text",
              content: "Column 1 text",
            },
          ],
        },
        {
          type: "column",
          blocks: [
            {
              type: "text",
              content: "Column 2 text",
            },
            {
              type: "button",
              label: "Edit me",
            },
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
      ],
    },
  ],
} satisfies DirektivPagesType;

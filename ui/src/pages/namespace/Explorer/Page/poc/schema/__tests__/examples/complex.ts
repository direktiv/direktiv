import { DirektivPagesType } from "../..";

export default {
  direktiv_api: "pages/v1",
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
              content: "Some text goes here",
            },
          ],
        },
        {
          type: "column",
          blocks: [
            {
              type: "query-provider",
              queries: [
                {
                  id: "fetching-resources",
                  endpoint: "/api/get/resources",
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
        },
      ],
    },
  ],
} satisfies DirektivPagesType;

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
      columns: [
        [
          {
            type: "text",
            content: "Some text goes here",
          },
        ],
        [
          {
            type: "query-provider",
            queries: [
              {
                id: "company-list",
                endpoint: "/api/get/companies",
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
                type: "table",
                data: {
                  type: "loop",
                  id: "company",
                  data: "query.company-list.data",
                  blocks: [],
                },
                actions: [
                  {
                    type: "button",
                    label: "delete",
                  },
                  {
                    type: "button",
                    label: "edit",
                  },
                ],
                columns: [
                  {
                    type: "table-column",
                    label: "name",
                    content: "{{loop.company.name}}",
                  },
                  {
                    type: "table-column",
                    label: "email",
                    content: "{{loop.company.email}}",
                  },
                ],
              },
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
      ],
    },
  ],
} satisfies DirektivPagesType;

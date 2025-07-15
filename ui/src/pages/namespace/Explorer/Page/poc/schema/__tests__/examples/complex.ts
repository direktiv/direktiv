import { DirektivPagesType } from "../..";

export default {
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
                id: "create-ticket",
                url: "/api/teams/{{query.user.teamId}}/projects/{{loop.project.id}}/tickets",
                method: "POST",
                queryParams: [
                  {
                    key: "assigned",
                    value: "{{query.user.id}}",
                  },
                ],
                requestHeaders: [
                  {
                    key: "Authorization",
                    value: "Bearer {{query.user.token}}",
                  },
                ],
                requestBody: [
                  {
                    key: "title",
                    value: {
                      type: "string",
                      value: "Draft: {{form.ticketForm.title}}",
                    },
                  },
                  {
                    key: "description",
                    value: {
                      type: "string",
                      value: "Steps to reproduce: \n\n Acceptance criteria: \n",
                    },
                  },
                  {
                    key: "priority",
                    value: {
                      type: "variable",
                      value: "form.ticketForm.priority",
                    },
                  },
                  {
                    key: "hidden",
                    value: {
                      type: "variable",
                      value: "form.ticketForm.hidden",
                    },
                  },
                  {
                    key: "isDraft",
                    value: {
                      type: "boolean",
                      value: true,
                    },
                  },
                  {
                    key: "categories",
                    value: {
                      type: "variable",
                      value: "loop.project.categories",
                    },
                  },
                  {
                    key: "relatedTickets",
                    value: {
                      type: "array",
                      value: ["ticket-1", "ticket-2", "ticket-3"],
                    },
                  },
                  {
                    key: "customFields",
                    value: {
                      type: "object",
                      value: [
                        {
                          key: "severity",
                          value: "high",
                        },
                        {
                          key: "environment",
                          value: "staging",
                        },
                      ],
                    },
                  },
                ],
              },
              blocks: [],
            },
          ],
        },
      ],
    },
  ],
} satisfies DirektivPagesType;

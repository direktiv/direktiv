import "../App.css";
import "../i18n";

import { Block } from "../pages/namespace/Explorer/Page/poc/PageCompiler/Block";
import { BlockList } from "../pages/namespace/Explorer/Page/poc/PageCompiler/Block/utils/BlockList";
import { DirektivPagesType } from "../pages/namespace/Explorer/Page/poc/schema";
import { EditorPanelLayoutProvider } from "../pages/namespace/Explorer/Page/poc/BlockEditor/EditorPanelProvider";
import { PageCompilerContextProvider } from "../pages/namespace/Explorer/Page/poc/PageCompiler/context/pageCompilerContext";
import { QueryClientProvider } from "@tanstack/react-query";
import React from "react";
import { Toaster } from "~/design/Toast";
import { createRoot } from "react-dom/client";
import queryClient from "../util/queryClient";
import { setPage } from "../pages/namespace/Explorer/Page/poc/PageCompiler/__tests__/utils";

const appContainer = document.getElementById("root");
if (!appContainer) throw new Error("Root element not found");

const page: DirektivPagesType = {
  direktiv_api: "page/v1",
  type: "page",
  blocks: [
    {
      type: "headline",
      level: "h1",
      label: "ThunderCorp International",
    },
    {
      type: "columns",
      blocks: [
        {
          type: "column",
          blocks: [
            {
              type: "image",
              src: "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/other/dream-world/25.svg",
              width: 200,
              height: 200,
            },
          ],
        },
        {
          type: "column",
          blocks: [
            {
              type: "headline",
              level: "h3",
              label: "CEO: Pikachu",
            },
            {
              type: "text",
              content:
                "Pikachu is the CEO of ThunderCorp International, known for his energetic leadership and iconic rise from pop culture mascot to business figure. With a shockingly sharp strategy, he leads with agility, loyalty, and charm.",
            },
            {
              type: "dialog",
              trigger: {
                type: "button",
                label: "Read more",
              },
              blocks: [
                {
                  type: "text",
                  content:
                    "Under Pikachu's leadership, ThunderCorp has expanded into cutting-edge sectors including sustainable energy, gamified wellness platforms, and AI-driven entertainment. His approach blends instinctive decision-making with deep emotional intelligence, earning him a reputation as both a visionary and a team-first executive. Despite his small stature, Pikachu commands attention in boardrooms worldwide, often communicating through subtle gestures, electrifying presence, and the occasional “Pika” punctuated with purpose. A firm believer in evolution—both personal and organizational—he champions growth through adaptation, partnership, and a relentless commitment to staying one step ahead in an ever-changing landscape.",
                },
              ],
            },
          ],
        },
      ],
    },
    {
      type: "query-provider",
      queries: [
        {
          id: "employees",
          url: "http://localhost:8080/employees",
          queryParams: [],
        },
        {
          id: "departments",
          url: "http://localhost:8080/departments",
          queryParams: [],
        },
      ],
      blocks: [
        {
          type: "headline",
          level: "h1",
          label:
            "{{query.employees.count}} employees in {{query.departments.count}} departments",
        },
        {
          type: "table",
          data: {
            type: "loop",
            id: "person",
            data: "query.employees.data",
            pageSize: 5,
          },
          blocks: [
            {
              type: "table-actions",
              blocks: [
                {
                  type: "dialog",
                  trigger: {
                    type: "button",
                    label: "new employee",
                  },
                  blocks: [
                    {
                      type: "headline",
                      level: "h3",
                      label: "Create new employee",
                    },
                    {
                      type: "form",
                      trigger: {
                        type: "button",
                        label: "Create",
                      },
                      mutation: {
                        method: "POST",
                        url: "http://localhost:8080/employees",
                        requestBody: [
                          {
                            key: "name",
                            value: {
                              type: "variable",
                              value: "this.name",
                            },
                          },
                          {
                            key: "age",
                            value: {
                              type: "variable",
                              value: "this.age",
                            },
                          },
                          {
                            key: "isActive",
                            value: {
                              type: "variable",
                              value: "this.active",
                            },
                          },
                          {
                            key: "department",
                            value: {
                              type: "variable",
                              value: "this.department",
                            },
                          },
                          {
                            key: "salary",
                            value: {
                              type: "variable",
                              value: "this.salary",
                            },
                          },
                        ],
                      },
                      blocks: [
                        {
                          id: "name",
                          label: "name",
                          description: "",
                          optional: false,
                          type: "form-string-input",
                          variant: "text",
                          defaultValue: "",
                        },
                        {
                          type: "columns",
                          blocks: [
                            {
                              type: "column",
                              blocks: [
                                {
                                  id: "salary",
                                  label: "salary",
                                  description: "",
                                  optional: false,
                                  type: "form-number-input",
                                  defaultValue: {
                                    type: "number",
                                    value: 0,
                                  },
                                },
                              ],
                            },
                            {
                              type: "column",
                              blocks: [
                                {
                                  id: "age",
                                  label: "age",
                                  description: "",
                                  optional: false,
                                  type: "form-number-input",
                                  defaultValue: {
                                    type: "number",
                                    value: 0,
                                  },
                                },
                              ],
                            },
                          ],
                        },
                        {
                          id: "department",
                          label: "department",
                          description: "",
                          optional: false,
                          type: "form-select",
                          values: {
                            type: "variable",
                            value: "query.departments.data",
                          },
                          defaultValue: "",
                        },
                        {
                          id: "active",
                          label: "active",
                          description: "This employee is active",
                          optional: true,
                          type: "form-checkbox",
                          defaultValue: {
                            type: "boolean",
                            value: false,
                          },
                        },
                      ],
                    },
                  ],
                },
                {
                  type: "dialog",
                  trigger: {
                    type: "button",
                    label: "import employees",
                  },
                  blocks: [
                    {
                      type: "headline",
                      level: "h3",
                      label: "Create random entries",
                    },
                    {
                      type: "form",
                      trigger: {
                        type: "button",
                        label: "Create entries",
                      },
                      mutation: {
                        method: "POST",
                        url: "http://localhost:8080/actions",
                        queryParams: [],
                        requestHeaders: [],
                        requestBody: [
                          {
                            key: "action",
                            value: {
                              type: "string",
                              value: "seed",
                            },
                          },
                          {
                            key: "count",
                            value: {
                              type: "variable",
                              value: "this.count",
                            },
                          },
                        ],
                      },
                      blocks: [
                        {
                          id: "count",
                          label: "Number of entries",
                          description: "",
                          optional: false,
                          type: "form-number-input",
                          defaultValue: {
                            type: "number",
                            value: 10,
                          },
                        },
                      ],
                    },
                  ],
                },
                {
                  type: "dialog",
                  trigger: {
                    type: "button",
                    label: "empty database",
                  },
                  blocks: [
                    {
                      type: "headline",
                      level: "h3",
                      label: "Empty database",
                    },
                    {
                      type: "form",
                      trigger: {
                        type: "button",
                        label: "empty database",
                      },
                      mutation: {
                        method: "POST",
                        url: "http://localhost:8080/actions",
                        requestBody: [
                          {
                            key: "action",
                            value: {
                              type: "string",
                              value: "clear",
                            },
                          },
                        ],
                      },
                      blocks: [
                        {
                          type: "text",
                          content: "Do you want to empty the database?",
                        },
                      ],
                    },
                  ],
                },
              ],
            },
            {
              type: "row-actions",
              blocks: [
                {
                  type: "dialog",
                  trigger: {
                    type: "button",
                    label: "open",
                  },
                  blocks: [
                    {
                      type: "headline",
                      level: "h3",
                      label: "{{loop.person.name}}",
                    },
                    {
                      type: "query-provider",
                      queries: [
                        {
                          id: "employee",
                          url: "http://localhost:8080/employees/{{loop.person.id}}",
                          queryParams: [],
                        },
                      ],
                      blocks: [
                        {
                          type: "columns",
                          blocks: [
                            {
                              type: "column",
                              blocks: [
                                {
                                  type: "image",
                                  src: "{{query.employee.data.image}}",
                                  width: 200,
                                  height: 200,
                                },
                              ],
                            },
                            {
                              type: "column",
                              blocks: [
                                {
                                  type: "text",
                                  content:
                                    "Salary: {{query.employee.data.salary}}",
                                },
                                {
                                  type: "text",
                                  content: "Age: {{query.employee.data.age}}",
                                },
                                {
                                  type: "text",
                                  content:
                                    "Department: {{query.employee.data.department}}",
                                },
                                {
                                  type: "text",
                                  content:
                                    "Active: {{query.employee.data.isActive}}",
                                },
                              ],
                            },
                          ],
                        },
                      ],
                    },
                  ],
                },
                {
                  type: "dialog",
                  trigger: {
                    type: "button",
                    label: "edit",
                  },
                  blocks: [
                    {
                      type: "headline",
                      level: "h3",
                      label: "Edit {{loop.person.name}}",
                    },
                    {
                      type: "query-provider",
                      queries: [
                        {
                          id: "employee",
                          url: "http://localhost:8080/employees/{{loop.person.id}}",
                          queryParams: [],
                        },
                      ],
                      blocks: [
                        {
                          type: "form",
                          trigger: {
                            type: "button",
                            label: "Update",
                          },
                          mutation: {
                            method: "PUT",
                            url: "http://localhost:8080/employees/{{loop.person.id}}",
                            requestBody: [
                              {
                                key: "name",
                                value: {
                                  type: "variable",
                                  value: "this.name",
                                },
                              },
                            ],
                          },
                          blocks: [
                            {
                              id: "name",
                              label: "name",
                              description: "",
                              optional: false,
                              type: "form-string-input",
                              variant: "text",
                              defaultValue: "{{query.employee.data.name}}",
                            },
                            {
                              type: "columns",
                              blocks: [
                                {
                                  type: "column",
                                  blocks: [
                                    {
                                      id: "salary",
                                      label: "salary",
                                      description: "",
                                      optional: false,
                                      type: "form-number-input",
                                      defaultValue: {
                                        type: "variable",
                                        value: "query.employee.data.salary",
                                      },
                                    },
                                  ],
                                },
                                {
                                  type: "column",
                                  blocks: [
                                    {
                                      id: "age",
                                      label: "age",
                                      description: "",
                                      optional: false,
                                      type: "form-number-input",
                                      defaultValue: {
                                        type: "variable",
                                        value: "query.employee.data.age",
                                      },
                                    },
                                  ],
                                },
                              ],
                            },
                            {
                              id: "department",
                              label: "department",
                              description: "",
                              optional: false,
                              type: "form-select",
                              values: {
                                type: "variable",
                                value: "query.departments.data",
                              },
                              defaultValue:
                                "{{query.employee.data.department}}",
                            },
                            {
                              id: "active",
                              label: "active",
                              description: "This employee is active",
                              optional: true,
                              type: "form-checkbox",
                              defaultValue: {
                                type: "variable",
                                value: "query.employee.data.isActive",
                              },
                            },
                          ],
                        },
                      ],
                    },
                  ],
                },
                {
                  type: "dialog",
                  trigger: {
                    type: "button",
                    label: "delete",
                  },
                  blocks: [
                    {
                      type: "headline",
                      level: "h3",
                      label: "Delete {{loop.person.name}}",
                    },
                    {
                      type: "form",
                      trigger: {
                        type: "button",
                        label: "Delete",
                      },
                      mutation: {
                        method: "DELETE",
                        url: "http://localhost:8080/employees/{{loop.person.id}}",
                      },
                      blocks: [
                        {
                          type: "text",
                          content:
                            "Are you sure you want to delete this employee?",
                        },
                      ],
                    },
                  ],
                },
              ],
            },
          ],
          columns: [
            {
              type: "table-column",
              label: "name",
              content: "{{loop.person.name}}",
            },
            {
              type: "table-column",
              label: "department",
              content: "{{loop.person.department}}",
            },
          ],
        },
      ],
    },
  ],
};

createRoot(appContainer).render(
  <React.StrictMode>
    <PageCompilerContextProvider setPage={setPage} page={page} mode="live">
      <QueryClientProvider client={queryClient}>
        <EditorPanelLayoutProvider>
          <BlockList path={[]}>
            {page.blocks.map((block, index) => (
              <Block key={index} block={block} blockPath={[index]} />
            ))}
          </BlockList>
        </EditorPanelLayoutProvider>
      </QueryClientProvider>
      <Toaster />
    </PageCompilerContextProvider>
  </React.StrictMode>
);

import { HttpResponse, http } from "msw";
import { act, render, screen } from "@testing-library/react";
import {
  afterAll,
  afterEach,
  beforeAll,
  describe,
  expect,
  test,
  vi,
} from "vitest";
import {
  getClientDetailsResponse,
  getCompanyListResponse,
} from "./api/samples";

import { PageCompiler } from "..";
import { createPage } from "./utils";
import { setupServer } from "msw/node";

// vi.mock("react-i18next", () => ({
//   useTranslation: () => ({
//     t: (key: string) => key,
//   }),
// }));

const apiServer = setupServer(
  http.get("/companies", () => HttpResponse.json(getCompanyListResponse)),
  http.get("/client/101", () => HttpResponse.json(getClientDetailsResponse))
);

beforeAll(() => {
  apiServer.listen({ onUnhandledRequest: "error" });
});

afterAll(() => apiServer.close());

afterEach(() => {
  vi.clearAllMocks();
  apiServer.resetHandlers();
});

describe("VariableString", () => {
  describe("invalid placeholders", () => {
    test("will show an error when the variable has an invalid namespace", async () => {
      await act(async () => {
        render(
          <PageCompiler
            page={createPage([
              {
                type: "headline",
                size: "h1",
                label: "template string without id: {{thisDoesNotExist}}",
              },
            ])}
            mode="live"
          />
        );
      });

      expect(screen.getByRole("heading", { level: 1 }).textContent).toBe(
        "template string without id: thisDoesNotExist (namespaceInvalid)"
      );
    });

    test("will show an error when the variable has no id", async () => {
      await act(async () => {
        render(
          <PageCompiler
            page={createPage([
              {
                type: "headline",
                size: "h1",
                label: "template string without id: {{query}}",
              },
            ])}
            mode="live"
          />
        );
      });

      expect(screen.getByRole("heading", { level: 1 }).textContent).toBe(
        "template string without id: query (idUndefined)"
      );
    });

    test("will show an error when the variable has no pointer", async () => {
      await act(async () => {
        render(
          <PageCompiler
            page={createPage([
              {
                type: "headline",
                size: "h1",
                label: "template string without id: {{query.id}}",
              },
            ])}
            mode="live"
          />
        );
      });

      expect(screen.getByRole("heading", { level: 1 }).textContent).toBe(
        "template string without id: query.id (pointerUndefined)"
      );
    });

    test("will show an error when the variable will point to an undefined value", async () => {
      await act(async () => {
        render(
          <PageCompiler
            page={createPage([
              {
                type: "headline",
                size: "h1",
                label: "template string without id: {{query.id.nothing}}",
              },
            ])}
            mode="live"
          />
        );
      });

      expect(screen.getByRole("heading", { level: 1 }).textContent).toBe(
        "template string without id: query.id.nothing (NoStateForId)"
      );
    });
  });

  describe("displaying data from a query provider", () => {
    test("childs can access data from a query provider", async () => {
      await act(async () => {
        render(
          <PageCompiler
            page={createPage([
              {
                type: "headline",
                size: "h1",
                label:
                  "Query not accessible from parent: {{query.company-list.total}}",
              },
              {
                type: "query-provider",
                queries: [
                  {
                    id: "company-list",
                    endpoint: "/companies",
                  },
                ],
                blocks: [
                  {
                    type: "headline",
                    size: "h2",
                    label: "{{query.company-list.total}} companies",
                  },
                  {
                    type: "card",
                    blocks: [
                      {
                        type: "query-provider",
                        queries: [
                          {
                            id: "client",
                            endpoint: "/client/101",
                          },
                        ],
                        blocks: [
                          {
                            type: "headline",
                            size: "h3",
                            label:
                              "Acces name from a deeper child: {{query.company-list.data.0.name}}, access another query: {{query.client.data.email}}",
                          },
                        ],
                      },
                    ],
                  },
                ],
              },
            ])}
            mode="live"
          />
        );
      });

      expect(screen.getByRole("heading", { level: 1 }).textContent).toBe(
        "Query not accessible from parent: query.company-list.total (NoStateForId)"
      );

      expect(screen.getByRole("heading", { level: 2 }).textContent).toBe(
        "10 companies"
      );

      expect(screen.getByRole("heading", { level: 3 }).textContent).toBe(
        "Acces name from a deeper child: Wintheiser-Lebsack, access another query: marisol.reichert@example.com"
      );
    });

    test("only data that can be stringified is accessible from a query provider", async () => {
      await act(async () => {
        render(
          <PageCompiler
            page={createPage([
              {
                type: "query-provider",
                queries: [
                  {
                    id: "company-list",
                    endpoint: "/companies",
                  },
                ],
                blocks: [
                  {
                    type: "headline",
                    size: "h1",
                    label: "Array does not work: {{query.company-list.data}}",
                  },
                  {
                    type: "headline",
                    size: "h2",
                    label:
                      "Object does not work: {{query.company-list.data.0}}",
                  },
                ],
              },
            ])}
            mode="live"
          />
        );
      });

      expect(screen.getByRole("heading", { level: 1 }).textContent).toBe(
        "Array does not work: query.company-list.data (couldNotStringify)"
      );

      expect(screen.getByRole("heading", { level: 2 }).textContent).toBe(
        "Object does not work: query.company-list.data.0 (couldNotStringify)"
      );
    });

    test("can not reuse a query id in one tree branch", async () => {
      await act(async () => {
        render(
          <PageCompiler
            page={createPage([
              {
                type: "query-provider",
                queries: [
                  {
                    id: "company-list",
                    endpoint: "/companies",
                  },
                ],
                blocks: [
                  {
                    type: "query-provider",
                    queries: [
                      {
                        id: "company-list",
                        endpoint: "/companies",
                      },
                    ],
                    blocks: [],
                  },
                ],
              },
            ])}
            mode="live"
          />
        );
      });

      expect(screen.getByRole("heading", { level: 2 }).textContent).toBe(
        "Object does not work: query.company-list.data.0 (couldNotStringify)"
      );
    });
  });

  // describe("looping over data from a query provider", () => {
  //   test("childs can access data from a query provider", async () => {
  //     await act(async () => {
  //       render(
  //         <PageCompiler
  //           page={createPage([
  //             {
  //               type: "headline",
  //               size: "h1",
  //               label:
  //                 "Query not accessible from parent: {{query.company-list.total}}",
  //             },
  //             {
  //               type: "query-provider",
  //               queries: [
  //                 {
  //                   id: "company-list",
  //                   endpoint: "/companies",
  //                 },
  //               ],
  //               blocks: [
  //                 {
  //                   type: "headline",
  //                   size: "h2",
  //                   label: "{{query.company-list.total}} companies",
  //                 },
  //                 {
  //                   type: "card",
  //                   blocks: [
  //                     {
  //                       type: "headline",
  //                       size: "h3",
  //                       label:
  //                         "Acces name from a deeper child: {{query.company-list.data.0.name}}",
  //                     },
  //                   ],
  //                 },
  //               ],
  //             },
  //           ])}
  //           mode="live"
  //         />
  //       );
  //     });

  //     expect(screen.getByRole("heading", { level: 1 }).textContent).toBe(
  //       "Query not accessible from parent: query.company-list.total (NoStateForId)"
  //     );

  //     expect(screen.getByRole("heading", { level: 2 }).textContent).toBe(
  //       "10 companies"
  //     );

  //     expect(screen.getByRole("heading", { level: 3 }).textContent).toBe(
  //       "Acces name from a deeper child: Wintheiser-Lebsack"
  //     );
  //   });
  // });
});

// TODO: using a query provider with the same id twice
// TODO: using a loop with the same id twice
// TODO: loops with dublicate id

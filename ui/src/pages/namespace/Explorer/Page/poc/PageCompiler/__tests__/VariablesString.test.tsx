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
  dataTypesResponse,
  getClientDetailsResponse,
  getCompanyListResponse,
} from "./utils/api/samples";

import { DirektivPagesType } from "../../schema";
import { PageCompiler } from "..";
import { createDirektivPage } from "./utils";
import { setupServer } from "msw/node";

const apiServer = setupServer(
  http.get("/companies", () => HttpResponse.json(getCompanyListResponse)),
  http.get("/client/101", () => HttpResponse.json(getClientDetailsResponse)),
  http.get("/data-types", () => HttpResponse.json(dataTypesResponse))
);

const setPage = (page: DirektivPagesType) => page;

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
    test("shows an error when the variable has an invalid namespace", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPage([
              {
                type: "headline",
                level: "h1",
                label: "template string without id: {{thisDoesNotExist}}",
              },
            ])}
            mode="live"
          />
        );
      });

      expect(screen.getByRole("alert").textContent).toBe(
        "thisDoesNotExist (namespaceInvalid)"
      );
    });

    test("shows an error when the variable has no id", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPage([
              {
                type: "headline",
                level: "h1",
                label: "template string without id: {{query}}",
              },
            ])}
            mode="live"
          />
        );
      });

      expect(screen.getByRole("alert").textContent).toBe("query (idUndefined)");
    });

    test("shows an error when the variable has no pointer", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPage([
              {
                type: "headline",
                level: "h1",
                label: "template string without id: {{query.id}}",
              },
            ])}
            mode="live"
          />
        );
      });

      expect(screen.getByRole("alert").textContent).toBe(
        "query.id (pointerUndefined)"
      );
    });

    test("shows an error when the variable points to an undefined id", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPage([
              {
                type: "headline",
                level: "h1",
                label: "template string without id: {{query.id.nothing}}",
              },
            ])}
            mode="live"
          />
        );
      });

      expect(screen.getByRole("alert").textContent).toBe(
        "query.id.nothing (NoStateForId)"
      );
    });

    test("shows an error when the variable points to an undefined value", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPage([
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
                    level: "h2",
                    label:
                      "{{query.company-list.this-does-not-exist}} companies",
                  },
                ],
              },
            ])}
            mode="live"
          />
        );
      });

      expect(screen.getByRole("alert").textContent).toBe(
        "query.company-list.this-does-not-exist (invalidPath)"
      );
    });
  });

  describe("display data from a query provider", () => {
    test("children can access data from a query provider", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPage([
              {
                type: "headline",
                level: "h1",
                label:
                  "Query cannot be reached from a parent: {{query.company-list.total}}",
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
                    level: "h2",
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
                            level: "h3",
                            label:
                              "Access name from a deeper child: {{query.company-list.data.0.name}}, access another query: {{query.client.data.email}}",
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
        "Query cannot be reached from a parent: query.company-list.total (NoStateForId)"
      );

      expect(screen.getByRole("heading", { level: 2 }).textContent).toBe(
        "10 companies"
      );

      expect(screen.getByRole("heading", { level: 3 }).textContent).toBe(
        "Access name from a deeper child: Wintheiser-Lebsack, access another query: marisol.reichert@example.com"
      );
    });

    test("only serializable data is accessible from a query provider", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPage([
              {
                type: "query-provider",
                queries: [
                  {
                    id: "data-types",
                    endpoint: "/data-types",
                  },
                ],
                blocks: [
                  {
                    type: "headline",
                    level: "h1",
                    label: "Array does not work: {{query.data-types.array}}",
                  },
                  {
                    type: "headline",
                    level: "h1",
                    label: "Object does not work: {{query.data-types.object}}",
                  },
                  {
                    type: "headline",
                    level: "h1",
                    label:
                      "undefiend does not work: {{query.data-types.object}}",
                  },
                  {
                    type: "headline",
                    level: "h2",
                    label: "string does work: {{query.data-types.string}}",
                  },
                  {
                    type: "headline",
                    level: "h2",
                    label: "boolean does work: {{query.data-types.boolean}}",
                  },
                  {
                    type: "headline",
                    level: "h2",
                    label: "null does work: {{query.data-types.null}}",
                  },
                  {
                    type: "headline",
                    level: "h2",
                    label: "number does work: {{query.data-types.number}}",
                  },
                ],
              },
            ])}
            mode="live"
          />
        );
      });

      expect(
        screen.getAllByRole("heading", { level: 1 }).map((el) => el.textContent)
      ).toEqual([
        "Array does not work: query.data-types.array (couldNotStringify)",
        "Object does not work: query.data-types.object (couldNotStringify)",
        "undefiend does not work: query.data-types.object (couldNotStringify)",
      ]);

      expect(
        screen.getAllByRole("heading", { level: 2 }).map((el) => el.textContent)
      ).toEqual([
        "string does work: hello world",
        "boolean does work: true",
        "null does work: null",
        "number does work: 123",
      ]);
    });

    test("reusing a query ID within the same branch is disallowed", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPage([
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

      expect(
        screen.getByLabelText("There was an unexpected error")
      ).toBeDefined();
    });
  });

  describe("loop over data from a query provider", () => {
    test("children can access data from a loop", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPage([
              {
                type: "headline",
                level: "h1",
                label:
                  "Loop cannot be reached from a parent: {{loop.company.name}}",
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
                    type: "card",
                    blocks: [
                      {
                        type: "loop",
                        id: "company",
                        data: "query.company-list.data",
                        blocks: [
                          {
                            type: "headline",
                            level: "h2",
                            label: "Company name: {{loop.company.name}}",
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
        "Loop cannot be reached from a parent: loop.company.name (NoStateForId)"
      );

      expect(
        screen.getAllByRole("heading", { level: 2 }).map((el) => el.textContent)
      ).toEqual([
        "Company name: Wintheiser-Lebsack",
        "Company name: Tremblay, Rohan and Friesen",
        "Company name: Wiegand Inc",
        "Company name: Champlin Group",
        "Company name: McCullough, Ziemann and Hirthe",
        "Company name: Hayes-Stracke",
        "Company name: Schroeder-Gleason",
        "Company name: Daugherty Inc",
        "Company name: Bartell, Champlin and Ziemann",
        "Company name: Quigley, Steuber and Gibson",
      ]);
    });

    test("children can access data from nested loops", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPage([
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
                            type: "loop",
                            id: "company",
                            data: "query.company-list.data",
                            blocks: [
                              {
                                type: "headline",
                                level: "h2",
                                label: "outer loop: {{loop.company.name}}",
                              },
                              {
                                type: "loop",
                                id: "clientAddress",
                                data: "query.client.data.addresses",
                                blocks: [
                                  {
                                    type: "headline",
                                    level: "h2",
                                    label:
                                      "-- inner loop: {{loop.company.id}} {{loop.clientAddress.city}}",
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
              },
            ])}
            mode="live"
          />
        );
      });

      expect(
        screen.getAllByRole("heading", { level: 2 }).map((el) => el.textContent)
      ).toEqual([
        "outer loop: Wintheiser-Lebsack",
        "-- inner loop: 1 East Alanamouth",
        "-- inner loop: 1 West Trevorview",
        "outer loop: Tremblay, Rohan and Friesen",
        "-- inner loop: 2 East Alanamouth",
        "-- inner loop: 2 West Trevorview",
        "outer loop: Wiegand Inc",
        "-- inner loop: 3 East Alanamouth",
        "-- inner loop: 3 West Trevorview",
        "outer loop: Champlin Group",
        "-- inner loop: 4 East Alanamouth",
        "-- inner loop: 4 West Trevorview",
        "outer loop: McCullough, Ziemann and Hirthe",
        "-- inner loop: 5 East Alanamouth",
        "-- inner loop: 5 West Trevorview",
        "outer loop: Hayes-Stracke",
        "-- inner loop: 6 East Alanamouth",
        "-- inner loop: 6 West Trevorview",
        "outer loop: Schroeder-Gleason",
        "-- inner loop: 7 East Alanamouth",
        "-- inner loop: 7 West Trevorview",
        "outer loop: Daugherty Inc",
        "-- inner loop: 8 East Alanamouth",
        "-- inner loop: 8 West Trevorview",
        "outer loop: Bartell, Champlin and Ziemann",
        "-- inner loop: 9 East Alanamouth",
        "-- inner loop: 9 West Trevorview",
        "outer loop: Quigley, Steuber and Gibson",
        "-- inner loop: 10 East Alanamouth",
        "-- inner loop: 10 West Trevorview",
      ]);
    });

    test("shows an error when loop data is not an array", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPage([
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
                    type: "card",
                    blocks: [
                      {
                        type: "loop",
                        id: "company",
                        data: "query.company-list.total",
                        blocks: [
                          {
                            type: "headline",
                            level: "h1",
                            label: "This will not be rendered",
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

      expect(screen.getByRole("alert").textContent).toBe(
        "query.company-list.total (notAnArray)"
      );

      expect(
        screen.queryByRole("heading", { level: 1 }),
        "Child blocks of the loop are not rendered"
      ).toBeNull();
    });

    test("reusing a loop ID within the same branch is disallowed", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPage([
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
                    type: "card",
                    blocks: [
                      {
                        type: "loop",
                        id: "company",
                        data: "query.company-list.data",
                        blocks: [
                          {
                            type: "loop",
                            id: "company",
                            data: "query.company-list.data",
                            blocks: [
                              {
                                type: "headline",
                                level: "h1",
                                label: "This will not be rendered",
                              },
                            ],
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

      expect(
        screen.getAllByLabelText("There was an unexpected error").length,
        "It renders the parsing error for every loop"
      ).toBe(10);

      expect(
        screen.queryByRole("heading", { level: 1 }),
        "Child blocks of the loop are not rendered"
      ).toBeNull();
    });
  });
});

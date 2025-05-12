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

  describe("display data from a query provider", () => {
    test("childs can access data from a query provider", async () => {
      await act(async () => {
        render(
          <PageCompiler
            page={createPage([
              {
                type: "headline",
                size: "h1",
                label:
                  "Query can not be reached from a parent: {{query.company-list.total}}",
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
        "Query can not be reached from a parent: query.company-list.total (NoStateForId)"
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

      // TODO: make this better, and fix i18n
      expect(screen.getByLabelText("parsing-error")).toBeDefined();
    });
  });

  // TODO: check if all errors from the translation file are covered
  // TODO: check if all screen. are as specific as possible
  describe("loop over data from a query provider", () => {
    test("children can access data from a loop", async () => {
      await act(async () => {
        render(
          <PageCompiler
            page={createPage([
              {
                type: "headline",
                size: "h1",
                label:
                  "Loop can not be reached from a parent: {{loop.company.name}}",
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
                            size: "h2",
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
        "Loop can not be reached from a parent: loop.company.name (NoStateForId)"
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
                                size: "h2",
                                label: "outer loop: {{loop.company.name}}",
                              },
                              {
                                type: "loop",
                                id: "clientAddress",
                                data: "query.client.data.addresses",
                                blocks: [
                                  {
                                    type: "headline",
                                    size: "h2",
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

    test("loop does not work with non array data", async () => {
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
                    type: "card",
                    blocks: [
                      {
                        type: "loop",
                        id: "company",
                        data: "query.company-list.total",
                        blocks: [
                          {
                            type: "headline",
                            size: "h1",
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

    test("can not reuse a loop id in one tree branch", async () => {
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
                                size: "h1",
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

      screen.debug();

      // TODO: make this better, and fix i18n
      expect(
        screen.getAllByLabelText("parsing-error").length,
        "It renders the parsing error for every loop"
      ).toBe(10);

      expect(
        screen.queryByRole("heading", { level: 1 }),
        "Child blocks of the loop are not rendered"
      ).toBeNull();
    });
  });
});

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

import { DirektivPagesType } from "../../schema";
import { PageCompiler } from "..";
import { createDirektivPage } from "./utils";
import { setupServer } from "msw/node";

const setPage = (page: DirektivPagesType) => page;

const apiRequestMock = vi.fn();

const apiServer = setupServer(
  http.get("/json-response", (...args) => {
    apiRequestMock(...args);
    return HttpResponse.json({
      data: {
        id: "1",
        message: "hello from the server",
      },
    });
  }),
  http.get("/dynamic/:id/path", (...args) => {
    apiRequestMock(...args);
    return HttpResponse.json({
      data: {
        message: "hello from the server",
      },
    });
  }),
  http.get("/text-response", (...args) => {
    apiRequestMock(...args);
    return HttpResponse.text("this is a text response");
  }),
  http.get("/404", (...args) => {
    apiRequestMock(...args);
    return HttpResponse.json({ error: "not found" }, { status: 404 });
  })
);

beforeAll(() => {
  apiServer.listen({ onUnhandledRequest: "error" });
});

afterAll(() => apiServer.close());

afterEach(() => {
  vi.clearAllMocks();
  apiServer.resetHandlers();
});

describe("QueryProvider", () => {
  test("makes the query result available to its child blocks", async () => {
    await act(async () => {
      render(
        <PageCompiler
          setPage={setPage}
          page={createDirektivPage([
            {
              type: "query-provider",
              queries: [
                {
                  id: "json-response",
                  url: "/json-response",
                },
              ],
              blocks: [
                {
                  type: "headline",
                  level: "h1",
                  label:
                    "This comes from the query provider: {{query.json-response.data.message}}",
                },
              ],
            },
          ])}
          mode="live"
        />
      );
    });

    expect(screen.getByRole("heading", { level: 1 }).textContent).toBe(
      "This comes from the query provider: hello from the server"
    );
  });

  test("will interpolate variables in the url and query params", async () => {
    await act(async () => {
      render(
        <PageCompiler
          setPage={setPage}
          page={createDirektivPage([
            {
              type: "query-provider",
              queries: [
                {
                  id: "json-response",
                  url: "/json-response",
                },
              ],
              blocks: [
                {
                  type: "query-provider",
                  queries: [
                    {
                      id: "request-with-variables",
                      url: "/dynamic/{{query.json-response.data.id}}/path",
                      queryParams: [
                        { key: "id", value: "{{query.json-response.data.id}}" },
                      ],
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

    expect(apiRequestMock).toHaveBeenCalledTimes(2);

    const [, secondRequest] = apiRequestMock.mock.calls;
    const [params] = secondRequest;
    const secondRequestUrl = new URL(params.request.url);

    expect(secondRequestUrl.pathname).toBe("/dynamic/1/path");
    expect(secondRequestUrl.search).toBe("?id=1");
  });

  test("shows an error when the query returns a status code outside of the 200 range", async () => {
    await act(async () => {
      render(
        <PageCompiler
          setPage={setPage}
          page={createDirektivPage([
            {
              type: "query-provider",
              queries: [
                {
                  id: "404",
                  url: "/404",
                },
              ],
              blocks: [
                {
                  type: "headline",
                  level: "h1",
                  label: "This will not be rendered",
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

    expect(
      screen.queryByRole("heading", { level: 1 }),
      "Child blocks of the query provider are not rendered"
    ).toBeNull();
  });

  test("shows an error when the query returns a non json response", async () => {
    await act(async () => {
      render(
        <PageCompiler
          setPage={setPage}
          page={createDirektivPage([
            {
              type: "query-provider",
              queries: [
                {
                  id: "text-response",
                  url: "/text-response",
                },
              ],
              blocks: [
                {
                  type: "headline",
                  level: "h1",
                  label: "This will not be rendered",
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

    expect(
      screen.queryByRole("heading", { level: 1 }),
      "Child blocks of the query provider are not rendered"
    ).toBeNull();
  });
});

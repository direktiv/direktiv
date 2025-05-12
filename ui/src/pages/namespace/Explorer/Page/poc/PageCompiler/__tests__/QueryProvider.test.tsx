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

import { PageCompiler } from "..";
import { createDirektivPage } from "./utils";
import { setupServer } from "msw/node";

const apiServer = setupServer(
  http.get("/json-response", () =>
    HttpResponse.json({
      data: { message: "hello from the server" },
    })
  ),
  http.get("/text-response", () =>
    HttpResponse.text("this is a text response")
  ),
  http.get("/404", () =>
    HttpResponse.json({ error: "not found" }, { status: 404 })
  )
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
          page={createDirektivPage([
            {
              type: "query-provider",
              queries: [
                {
                  id: "json-response",
                  endpoint: "/json-response",
                },
              ],
              blocks: [
                {
                  type: "headline",
                  size: "h1",
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

  test("shows an error when the query returns a status code outside of the 200 range", async () => {
    await act(async () => {
      render(
        <PageCompiler
          page={createDirektivPage([
            {
              type: "query-provider",
              queries: [
                {
                  id: "404",
                  endpoint: "/404",
                },
              ],
              blocks: [
                {
                  type: "headline",
                  size: "h1",
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
          page={createDirektivPage([
            {
              type: "query-provider",
              queries: [
                {
                  id: "text-response",
                  endpoint: "/text-response",
                },
              ],
              blocks: [
                {
                  type: "headline",
                  size: "h1",
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

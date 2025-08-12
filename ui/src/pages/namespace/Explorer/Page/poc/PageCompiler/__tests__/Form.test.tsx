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
import { createDirektivPage, setPage, setupResizeObserverMock } from "./utils";

import { BlockType } from "../../schema/blocks";
import { PageCompiler } from "..";
import { getUserDetailsResponse } from "./utils/api/samples";
import { setupServer } from "msw/node";

const apiServer = setupServer(
  http.get("/user-details", () => HttpResponse.json(getUserDetailsResponse))
);

beforeAll(() => {
  setupResizeObserverMock();
  apiServer.listen({ onUnhandledRequest: "error" });
});

afterAll(() => apiServer.close());

afterEach(() => {
  vi.clearAllMocks();
  apiServer.resetHandlers();
});

export const createForm = (blocks: BlockType[]) =>
  createDirektivPage([
    {
      type: "query-provider",
      queries: [
        {
          id: "user",
          url: "/user-details",
          queryParams: [],
        },
      ],
      blocks: [
        {
          type: "form",
          trigger: {
            type: "button",
            label: "form",
          },
          mutation: {
            id: "form",
            method: "POST",
            url: "/some-endpoint",
          },
          blocks,
        },
      ],
    },
  ]);

describe("Form", () => {
  describe("setting default values", () => {
    test("string input can use string templates in the default value attribute", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createForm([
              {
                id: "string",
                label: "string input",
                description: "",
                optional: false,
                type: "form-string-input",
                variant: "text",
                defaultValue:
                  "a string input can use variable placeholders like string:{{query.user.data.status}}, number: {{query.user.data.userId}} and booleans: {{query.user.data.emailVerified}}",
              },
            ])}
            mode="live"
          />
        );
      });

      expect((screen.getByRole("textbox") as HTMLInputElement)?.value).toBe(
        "a string input can use variable placeholders like string:ok, number: 101 and booleans: true"
      );
    });

    test("textarea can use string templates in the default value attribute", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createForm([
              {
                id: "string",
                label: "string input",
                description: "",
                optional: false,
                type: "form-textarea",
                defaultValue:
                  "a textarea can use variable placeholders like string:{{query.user.data.status}}, number: {{query.user.data.userId}} and booleans: {{query.user.data.emailVerified}}",
              },
            ])}
            mode="live"
          />
        );
      });

      expect((screen.getByRole("textbox") as HTMLInputElement)?.value).toBe(
        "a textarea can use variable placeholders like string:ok, number: 101 and booleans: true"
      );
    });

    test("checkbox can be checked by default", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createForm([
              {
                id: "static-checkbox",
                label: "static checkbox",
                description: "default values is always checked",
                optional: false,
                type: "form-checkbox",
                defaultValue: {
                  type: "boolean",
                  value: true,
                },
              },
            ])}
            mode="live"
          />
        );
      });
      expect(screen.getByRole("checkbox", { checked: true }));
    });

    test("checkbox can have a default value sourced from a variable", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createForm([
              {
                id: "static-checkbox",
                label: "static checkbox",
                description: "default values is always checked",
                optional: false,
                type: "form-checkbox",
                defaultValue: {
                  type: "variable",
                  value: "query.user.data.emailVerified",
                },
              },
            ])}
            mode="live"
          />
        );
      });
      expect(screen.getByRole("checkbox", { checked: true }));
    });
  });
});

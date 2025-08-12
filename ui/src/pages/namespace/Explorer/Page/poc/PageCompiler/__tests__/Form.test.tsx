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
import { createDirektivPage, setPage } from "./utils";

import { PageCompiler } from "..";
import { getUserDetailsResponse } from "./utils/api/samples";
import { setupServer } from "msw/node";

const apiServer = setupServer(
  http.get("/user-details", () => HttpResponse.json(getUserDetailsResponse))
);

beforeAll(() => {
  apiServer.listen({ onUnhandledRequest: "error" });
});

afterAll(() => apiServer.close());

afterEach(() => {
  vi.clearAllMocks();
  apiServer.resetHandlers();
});

describe("Form", () => {
  describe("setting default values", () => {
    test("string inputs and textareas can use string templates in the default value attribute", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPage([
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
                    blocks: [
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
                      {
                        id: "textarea",
                        label: "textarea",
                        description: "",
                        optional: false,
                        type: "form-textarea",
                        defaultValue:
                          "a textarea can use variable placeholders like string:{{query.user.data.status}}, number: {{query.user.data.userId}} and booleans: {{query.user.data.emailVerified}}",
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
        (screen.getAllByRole("textbox")[0] as HTMLInputElement)?.value
      ).toBe(
        "a string input can use variable placeholders like string:ok, number: 101 and booleans: true"
      );

      expect(
        (screen.getAllByRole("textbox")[1] as HTMLTextAreaElement)?.value
      ).toBe(
        "a textarea can use variable placeholders like string:ok, number: 101 and booleans: true"
      );
    });
  });
});

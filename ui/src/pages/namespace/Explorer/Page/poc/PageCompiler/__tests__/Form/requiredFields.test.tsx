import { act, render, screen, waitFor } from "@testing-library/react";
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
  createDirektivPageWithForm,
  setPage,
  setupResizeObserverMock,
} from "../utils";

import { PageCompiler } from "../..";
import { setupFormApi } from "./utils";

const { apiServer } = setupFormApi();

beforeAll(() => {
  setupResizeObserverMock();
  apiServer.listen({ onUnhandledRequest: "error" });
});

afterAll(() => apiServer.close());

afterEach(() => {
  vi.clearAllMocks();
  apiServer.resetHandlers();
});

describe("required fields", () => {
  describe("string input field", () => {
    test("shows an error when a required string input field is missing a value", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm(
              [
                {
                  id: "string",
                  label: "string input",
                  description: "",
                  optional: false,
                  type: "form-string-input",
                  variant: "text",
                  defaultValue: "",
                },
              ],
              {
                id: "save-user",
                method: "POST",
                url: "/save-user",
                requestBody: [],
              }
            )}
            mode="live"
          />
        );
      });

      await screen.getByRole("button", { name: "save" }).click();
      expect(screen.getAllByText("Some required fields are missing (string)"));
    });

    test("shows an error when a required string input field is missing a value", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm(
              [
                {
                  id: "string",
                  label: "string input",
                  description: "",
                  optional: false,
                  type: "form-string-input",
                  variant: "text",
                  defaultValue: "has default value",
                },
              ],
              {
                id: "save-user",
                method: "POST",
                url: "/save-user",
                requestBody: [],
              }
            )}
            mode="live"
          />
        );
      });

      await screen.getByRole("button", { name: "save" }).click();

      await waitFor(() => {
        expect(screen.getAllByText("The form has been submitted successfully"));
      });
    });
  });
});

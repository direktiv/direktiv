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

const { apiServer, apiRequestMock } = setupFormApi();

beforeAll(() => {
  setupResizeObserverMock();
  apiServer.listen({ onUnhandledRequest: "error" });
});

afterAll(() => apiServer.close());

afterEach(() => {
  vi.clearAllMocks();
  apiServer.resetHandlers();
});

describe("form request", () => {
  describe("request headers", () => {
    test("variables will be resolved and stringified", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm([], {
              id: "save-user",
              method: "POST",
              url: "/save-user",
              requestHeaders: [
                {
                  key: "String-Value",
                  value: "String: {{query.user.data.status}}",
                },
                {
                  key: "Boolean-Value",
                  value: "Boolean: {{query.user.data.emailVerified}}",
                },
                {
                  key: "Number-Value",
                  value: "Number: {{query.user.data.accountBalance}}",
                },
                {
                  key: "Null-Value",
                  value: "Null: {{query.user.data.lastLogin}}",
                },
              ],
            })}
            mode="live"
          />
        );
      });

      await screen.getByRole("button").click();

      await waitFor(() => {
        expect(apiRequestMock).toHaveBeenCalledTimes(1);
        const formRequest = apiRequestMock.mock.calls[0][0].request as Request;
        expect(formRequest.headers.get("String-Value")).toBe("String: ok");
        expect(formRequest.headers.get("Boolean-Value")).toBe("Boolean: true");
        expect(formRequest.headers.get("Number-Value")).toBe("Number: 19.99");
        expect(formRequest.headers.get("Null-Value")).toBe("Null: null");
      });
    });

    test("it shows an error when submitting a form that uses variables that can not be stringified", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm([], {
              id: "save-user",
              method: "POST",
              url: "/save-user",
              requestHeaders: [
                {
                  key: "object",
                  value: "String: {{query.user.data.profile}}",
                },
              ],
            })}
            mode="live"
          />
        );
      });

      await screen.getByRole("button").click();

      await waitFor(() => {
        expect(screen.getByRole("form").textContent).toContain(
          "Pointing to a value that can not be stringified. Make sure to point to either a String, Number, Boolean, or Null."
        );
      });
    });
  });
});

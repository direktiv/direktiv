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

import { BlockType } from "../../../schema/blocks";
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

const form: BlockType[] = [
  {
    id: "string",
    label: "string input",
    description: "",
    optional: false,
    type: "form-string-input",
    variant: "text",
    defaultValue: "string from a string input",
  },
  {
    id: "textarea",
    label: "textarea",
    description: "",
    optional: false,
    type: "form-textarea",
    defaultValue: "string from a textarea",
  },
  {
    id: "checkbox",
    label: "checkbox",
    description: "this is a checkbox",
    optional: false,
    type: "form-checkbox",
    defaultValue: {
      type: "boolean",
      value: true,
    },
  },
  {
    id: "number",
    label: "number input",
    description: "",
    optional: false,
    type: "form-number-input",
    defaultValue: {
      type: "number",
      value: 3,
    },
  },
  {
    id: "date",
    label: "date",
    description: "",
    optional: false,
    type: "form-date-input",
    defaultValue: "2025-12-24T00:00:00.000Z",
  },
  {
    id: "select",
    label: "select",
    description: "",
    optional: false,
    type: "form-select",
    values: ["free", "pro", "enterprise"],
    defaultValue: "pro",
  },
];

describe("form request", () => {
  describe("url", () => {
    test("variables will be resolved and stringified", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm(form, {
              id: "save-user",
              method: "POST",
              url: "/save-user",
              queryParams: [
                {
                  key: "string",
                  value: "{{query.user.data.status}}",
                },
                {
                  key: "boolean",
                  value: "{{query.user.data.emailVerified}}",
                },
                {
                  key: "number",
                  value: "{{query.user.data.accountBalance}}",
                },
                {
                  key: "null",
                  value: "{{query.user.data.lastLogin}}",
                },
                {
                  key: "form-string",
                  value: "{{form.save-user.string}}",
                },
                {
                  key: "form-textarea",
                  value: "{{form.save-user.textarea}}",
                },
                {
                  key: "checkbox",
                  value: "{{form.save-user.checkbox}}",
                },
                {
                  key: "number",
                  value: "{{form.save-user.number}}",
                },
                {
                  key: "date",
                  value: "{{form.save-user.date}}",
                },
                {
                  key: "select",
                  value: "{{form.save-user.select}}",
                },
              ],
            })}
            mode="live"
          />
        );
      });

      await screen.getByRole("button", { name: "save" }).click();

      await waitFor(() => {
        expect(apiRequestMock).toHaveBeenCalledTimes(1);
        const formRequest = apiRequestMock.mock.calls[0][0].request as Request;
        const requestUrl = new URL(formRequest.url);
        expect(requestUrl.search).toBe(
          // TODO: checkbox must evaluate to true and false
          "?string=ok&boolean=true&number=3&null=null&form-string=string+from+a+string+input&form-textarea=string+from+a+textarea&checkbox=on&date=2025-12-24&select=pro"
        );
      });
    });

    test("it shows an error when submitting a form that uses variables that can not be stringified", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm(form, {
              id: "save-user",
              method: "POST",
              url: "/save-user",
              queryParams: [
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

      await screen.getByRole("button", { name: "save" }).click();

      await waitFor(() => {
        expect(screen.getByRole("form").textContent).toContain(
          "Pointing to a value that can not be stringified. Make sure to point to either a String, Number, Boolean, or Null."
        );
      });
    });
  });

  describe("headers", () => {
    test("variables will be resolved and stringified", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm(form, {
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

      await screen.getByRole("button", { name: "save" }).click();

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
            page={createDirektivPageWithForm(form, {
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

      await screen.getByRole("button", { name: "save" }).click();

      await waitFor(() => {
        expect(screen.getByRole("form").textContent).toContain(
          "Pointing to a value that can not be stringified. Make sure to point to either a String, Number, Boolean, or Null."
        );
      });
    });
  });
});

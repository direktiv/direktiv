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
    id: "checkbox-checked",
    label: "checkbox",
    description: "this is a checked checkbox",
    optional: false,
    type: "form-checkbox",
    defaultValue: {
      type: "boolean",
      value: true,
    },
  },
  {
    id: "checkbox-unchecked",
    label: "checkbox",
    description: "this is a unchecked checkbox",
    optional: false,
    type: "form-checkbox",
    defaultValue: {
      type: "boolean",
      value: false,
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
    test("resolves variables in URL path", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm(form, {
              id: "delete-blog-post",
              method: "DELETE",
              url: "/blog-post/{{query.user.data.userId}}/{{this.form.string}}",
            })}
            mode="live"
          />
        );
      });

      await screen.getByRole("button", { name: "save" }).click();

      await waitFor(() => {
        expect(apiRequestMock).toHaveBeenCalledTimes(1);
        const formRequest = apiRequestMock.mock.calls[0][0].request as Request;
        const requestUrl = new URL(formRequest.clone().url);
        expect(requestUrl.pathname).toBe(
          "/blog-post/101/string%20from%20a%20string%20input"
        );
      });
    });

    test("shows error for non-stringifiable variables in URL", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm(form, {
              id: "save-user",
              method: "POST",
              url: "/save-user/{{query.user.data.profile}}",
            })}
            mode="live"
          />
        );
      });

      await screen.getByRole("button", { name: "save" }).click();

      await waitFor(() => {
        expect(screen.getByTestId("toast-error").textContent).toContain(
          "Variable error (query.user.data.profile): Pointing to a value that can not be stringified. Make sure to point to either a String, Number, Boolean, or Null."
        );
      });
    });
  });

  describe("query params", () => {
    test("resolves all variable types in query parameters", async () => {
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
                  value: "{{this.form.string}}",
                },
                {
                  key: "form-textarea",
                  value: "{{this.form.textarea}}",
                },
                {
                  key: "checkbox-checked",
                  value: "{{this.form.checkbox-checked}}",
                },
                {
                  key: "checkbox-unchecked",
                  value: "{{this.form.checkbox-unchecked}}",
                },
                {
                  key: "number",
                  value: "{{this.form.number}}",
                },
                {
                  key: "date",
                  value: "{{this.form.date}}",
                },
                {
                  key: "select",
                  value: "{{this.form.select}}",
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
        const requestUrl = new URL(formRequest.clone().url);
        expect(requestUrl.search).toBe(
          "?string=ok&boolean=true&number=3&null=null&form-string=string+from+a+string+input&form-textarea=string+from+a+textarea&checkbox-checked=true&checkbox-unchecked=false&date=2025-12-24&select=pro"
        );
      });
    });

    test("shows error for non-stringifiable variables in query params", async () => {
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
        expect(screen.getByTestId("toast-error").textContent).toContain(
          "Variable error (query.user.data.profile): Pointing to a value that can not be stringified. Make sure to point to either a String, Number, Boolean, or Null."
        );
      });
    });
  });

  describe("request headers", () => {
    test("resolves all variable types in request headers", async () => {
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
                  key: "Value",
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
                {
                  key: "Form-String-Value",
                  value: "String: {{this.form.string}}",
                },
                {
                  key: "Form-Textarea-Value",
                  value: "Textarea: {{this.form.textarea}}",
                },
                {
                  key: "Form-Checkbox-Checked-Value",
                  value: "Checkbox: {{this.form.checkbox-checked}}",
                },
                {
                  key: "Form-Checkbox-Unchecked-Value",
                  value: "Checkbox: {{this.form.checkbox-unchecked}}",
                },
                {
                  key: "Form-Number-Value",
                  value: "Number: {{this.form.number}}",
                },
                {
                  key: "Form-Date-Value",
                  value: "Date: {{this.form.date}}",
                },
                {
                  key: "Form-Select-Value",
                  value: "Select: {{this.form.select}}",
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
        expect(formRequest.clone().headers.get("Value")).toBe("String: ok");
        expect(formRequest.clone().headers.get("Boolean-Value")).toBe(
          "Boolean: true"
        );
        expect(formRequest.clone().headers.get("Number-Value")).toBe(
          "Number: 19.99"
        );
        expect(formRequest.clone().headers.get("Null-Value")).toBe(
          "Null: null"
        );
        expect(formRequest.clone().headers.get("Form-String-Value")).toBe(
          "String: string from a string input"
        );
        expect(formRequest.clone().headers.get("Form-Textarea-Value")).toBe(
          "Textarea: string from a textarea"
        );
        expect(
          formRequest.clone().headers.get("Form-Checkbox-Checked-Value")
        ).toBe("Checkbox: true");
        expect(
          formRequest.clone().headers.get("Form-Checkbox-Unchecked-Value")
        ).toBe("Checkbox: false");
        expect(formRequest.clone().headers.get("Form-Number-Value")).toBe(
          "Number: 3"
        );
        expect(formRequest.clone().headers.get("Form-Date-Value")).toBe(
          "Date: 2025-12-24"
        );
        expect(formRequest.clone().headers.get("Form-Select-Value")).toBe(
          "Select: pro"
        );
      });
    });

    test("shows error for non-stringifiable variables in headers", async () => {
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
        expect(screen.getByTestId("toast-error").textContent).toContain(
          "Variable error (query.user.data.profile): Pointing to a value that can not be stringified. Make sure to point to either a String, Number, Boolean, or Null."
        );
      });
    });
  });

  describe("request body", () => {
    test("resolves string types in request body", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm(form, {
              id: "save-user",
              method: "POST",
              url: "/save-user",
              requestBody: [
                {
                  key: "string-from-query",
                  value: {
                    type: "string",
                    value: "{{query.user.data.status}}",
                  },
                },
                {
                  key: "string-from-textarea",
                  value: {
                    type: "string",
                    value: "{{this.form.textarea}}",
                  },
                },
                {
                  key: "string-from-input",
                  value: {
                    type: "string",
                    value: "{{this.form.string}}",
                  },
                },
                {
                  key: "string-from-date-input",
                  value: {
                    type: "string",
                    value: "{{this.form.date}}",
                  },
                },
                {
                  key: "string-from-select",
                  value: {
                    type: "string",
                    value: "{{this.form.select}}",
                  },
                },
                {
                  key: "number-from-query",
                  value: {
                    type: "string",
                    value: "{{query.user.data.accountBalance}}",
                  },
                },
                {
                  key: "number-from-input",
                  value: {
                    type: "string",
                    value: "{{this.form.number}}",
                  },
                },
                {
                  key: "null-from-query",
                  value: {
                    type: "string",
                    value: "{{query.user.data.lastLogin}}",
                  },
                },
                {
                  key: "true-from-query",
                  value: {
                    type: "string",
                    value: "{{query.user.data.emailVerified}}",
                  },
                },
                {
                  key: "true-from-checkbox",
                  value: {
                    type: "string",
                    value: "{{this.form.checkbox-checked}}",
                  },
                },
                {
                  key: "false-from-checkbox",
                  value: {
                    type: "string",
                    value: "{{this.form.checkbox-unchecked}}",
                  },
                },
              ],
            })}
            mode="live"
          />
        );
      });

      await screen.getByRole("button", { name: "save" }).click();

      await waitFor(async () => {
        expect(apiRequestMock).toHaveBeenCalledTimes(1);
        const formRequest = apiRequestMock.mock.calls[0][0].request as Request;
        const text = JSON.parse(await formRequest.clone().text());

        expect(text["string-from-query"]).toBe("ok");
        expect(text["string-from-textarea"]).toBe("string from a textarea");
        expect(text["string-from-input"]).toBe("string from a string input");
        expect(text["string-from-date-input"]).toBe("2025-12-24");
        expect(text["string-from-select"]).toBe("pro");
        expect(text["number-from-query"]).toBe("19.99");
        expect(text["number-from-input"]).toBe("3");
        expect(text["null-from-query"]).toBe("null");
        expect(text["true-from-query"]).toBe("true");
        expect(text["true-from-checkbox"]).toBe("true");
        expect(text["false-from-checkbox"]).toBe("false");
      });
    });

    test("resolves variable types in request body", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm(form, {
              id: "save-user",
              method: "POST",
              url: "/save-user",
              requestBody: [
                {
                  key: "true-from-query",
                  value: {
                    type: "variable",
                    value: "query.user.data.emailVerified",
                  },
                },
                {
                  key: "true-from-checkbox",
                  value: {
                    type: "variable",
                    value: "this.form.checkbox-checked",
                  },
                },
                {
                  key: "false-from-checkbox",
                  value: {
                    type: "variable",
                    value: "this.form.checkbox-unchecked",
                  },
                },
                {
                  key: "number-from-query",
                  value: {
                    type: "variable",
                    value: "query.user.data.accountBalance",
                  },
                },

                {
                  key: "number-from-input",
                  value: {
                    type: "variable",
                    value: "this.form.number",
                  },
                },
                {
                  key: "array-from-query",
                  value: {
                    type: "variable",
                    value: "query.user.meta.subscriptionPlanOptions",
                  },
                },
                {
                  key: "null-from-query",
                  value: {
                    type: "variable",
                    value: "query.user.data.lastLogin",
                  },
                },
                {
                  key: "object-from-query",
                  value: {
                    type: "variable",
                    value: "query.user.data",
                  },
                },
              ],
            })}
            mode="live"
          />
        );
      });

      await screen.getByRole("button", { name: "save" }).click();

      await waitFor(async () => {
        expect(apiRequestMock).toHaveBeenCalledTimes(1);
        const formRequest = apiRequestMock.mock.calls[0][0].request as Request;
        const text = JSON.parse(await formRequest.clone().text());

        expect(text["true-from-query"]).toBe(true);
        expect(text["true-from-checkbox"]).toBe(true);
        expect(text["false-from-checkbox"]).toBe(false);
        expect(text["number-from-query"]).toBe(19.99);
        expect(text["number-from-input"]).toBe(3);
        expect(text["array-from-query"]).toEqual([
          { label: "Free Plan", value: "free" },
          { label: "Pro Plan", value: "pro" },
          { label: "Enterprise Plan", value: "enterprise" },
        ]);
        expect(text["null-from-query"]).toBe(null);
        expect(text["object-from-query"]).toEqual({
          accountBalance: 19.99,
          emailVerified: true,
          lastLogin: null,
          membershipStartDate: "2023-06-15T09:30:00Z",
          profile: { firstName: "Alice" },
          recentActivity: [
            "login",
            1623456789,
            false,
            null,
            { action: "purchase" },
          ],
          status: "ok",
          subscriptionPlan: "pro",
          twoFactorEnabled: false,
          userId: 101,
        });
      });
    });

    test("resolves boolean types in request body", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm(form, {
              id: "save-user",
              method: "POST",
              url: "/save-user",
              requestBody: [
                {
                  key: "true",
                  value: {
                    type: "boolean",
                    value: true,
                  },
                },
                {
                  key: "false",
                  value: {
                    type: "boolean",
                    value: false,
                  },
                },
              ],
            })}
            mode="live"
          />
        );
      });

      await screen.getByRole("button", { name: "save" }).click();

      await waitFor(async () => {
        expect(apiRequestMock).toHaveBeenCalledTimes(1);
        const formRequest = apiRequestMock.mock.calls[0][0].request as Request;
        const text = JSON.parse(await formRequest.clone().text());

        expect(text["true"]).toBe(true);
        expect(text["false"]).toBe(false);
      });
    });

    test("resolves number types in request body", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm(form, {
              id: "save-user",
              method: "POST",
              url: "/save-user",
              requestBody: [
                {
                  key: "number",
                  value: {
                    type: "number",
                    value: 3,
                  },
                },
              ],
            })}
            mode="live"
          />
        );
      });

      await screen.getByRole("button", { name: "save" }).click();

      await waitFor(async () => {
        expect(apiRequestMock).toHaveBeenCalledTimes(1);
        const formRequest = apiRequestMock.mock.calls[0][0].request as Request;
        const text = JSON.parse(await formRequest.clone().text());

        expect(text["number"]).toBe(3);
      });
    });

    test("resolves object types in request body", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm(form, {
              id: "save-user",
              method: "POST",
              url: "/save-user",
              requestBody: [
                {
                  key: "object",
                  value: {
                    type: "object",
                    value: [
                      {
                        key: "string",
                        value: "hello",
                      },
                      {
                        key: "number",
                        value: 1,
                      },
                      {
                        key: "boolean",
                        value: true,
                      },
                    ],
                  },
                },
              ],
            })}
            mode="live"
          />
        );
      });

      await screen.getByRole("button", { name: "save" }).click();

      await waitFor(async () => {
        expect(apiRequestMock).toHaveBeenCalledTimes(1);
        const formRequest = apiRequestMock.mock.calls[0][0].request as Request;
        const text = JSON.parse(await formRequest.clone().text());

        expect(text["object"]).toEqual({
          string: "hello",
          number: 1,
          boolean: true,
        });
      });
    });

    test("resolves array types in request body", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm(form, {
              id: "save-user",
              method: "POST",
              url: "/save-user",
              requestBody: [
                {
                  key: "string-array",
                  value: {
                    type: "string-array",
                    value: ["a", "b", "c"],
                  },
                },
                {
                  key: "boolean-array",
                  value: {
                    type: "boolean-array",
                    value: [true, false, false],
                  },
                },
                {
                  key: "number-array",
                  value: {
                    type: "number-array",
                    value: [1, 2, 3],
                  },
                },
              ],
            })}
            mode="live"
          />
        );
      });

      await screen.getByRole("button", { name: "save" }).click();

      await waitFor(async () => {
        expect(apiRequestMock).toHaveBeenCalledTimes(1);
        const formRequest = apiRequestMock.mock.calls[0][0].request as Request;
        const text = JSON.parse(await formRequest.clone().text());

        expect(text["string-array"]).toEqual(["a", "b", "c"]);
        expect(text["boolean-array"]).toEqual([true, false, false]);
        expect(text["number-array"]).toEqual([1, 2, 3]);
      });
    });

    test("shows error for non-serializable veriables in request body", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm(form, {
              id: "save-user",
              method: "POST",
              url: "/save-user",
              requestBody: [
                {
                  key: "object",
                  value: {
                    type: "string",
                    value: "String: {{query.user.data.profile}}",
                  },
                },
              ],
            })}
            mode="live"
          />
        );
      });

      await screen.getByRole("button", { name: "save" }).click();
      await waitFor(() => {
        expect(screen.getByTestId("toast-error").textContent).toContain(
          "Variable error (query.user.data.profile): Pointing to a value that can not be stringified. Make sure to point to either a String, Number, Boolean, or Null."
        );
      });
    });
  });
});

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
    id: "floating-number",
    label: "number input",
    description: "",
    optional: false,
    type: "form-number-input",
    defaultValue: {
      type: "number",
      value: 4.99,
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
                  key: "input",
                  value: "{{this.form.string}}",
                },
                {
                  key: "textarea",
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
                  key: "input-number",
                  value: "{{this.form.number}}",
                },
                {
                  key: "input-floating-number",
                  value: "{{this.form.floating-number}}",
                },
                {
                  key: "input-date",
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
          "?string=ok&boolean=true&number=19.99&null=null&input=string+from+a+string+input&textarea=string+from+a+textarea&checkbox-checked=true&checkbox-unchecked=false&input-number=3&input-floating-number=4.99&input-date=2025-12-24&select=pro"
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
                  key: "String",
                  value: "String: {{query.user.data.status}}",
                },
                {
                  key: "Boolean",
                  value: "Boolean: {{query.user.data.emailVerified}}",
                },
                {
                  key: "Number",
                  value: "Number: {{query.user.data.accountBalance}}",
                },
                {
                  key: "Null",
                  value: "Null: {{query.user.data.lastLogin}}",
                },
                {
                  key: "Input",
                  value: "Input: {{this.form.string}}",
                },
                {
                  key: "Textarea",
                  value: "Textarea: {{this.form.textarea}}",
                },
                {
                  key: "Checkbox-Checked",
                  value: "Checkbox (checked): {{this.form.checkbox-checked}}",
                },
                {
                  key: "Checkbox-Unchecked",
                  value:
                    "Checkbox (unchecked): {{this.form.checkbox-unchecked}}",
                },
                {
                  key: "Input-Number",
                  value: "Input Number: {{this.form.number}}",
                },
                {
                  key: "Input-Floating-Number",
                  value: "Input Floating Number: {{this.form.floating-number}}",
                },
                {
                  key: "Input-Date",
                  value: "Input Date: {{this.form.date}}",
                },
                {
                  key: "Select",
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
        expect(formRequest.clone().headers.get("String")).toBe("String: ok");
        expect(formRequest.clone().headers.get("Boolean")).toBe(
          "Boolean: true"
        );
        expect(formRequest.clone().headers.get("Number")).toBe("Number: 19.99");
        expect(formRequest.clone().headers.get("Null")).toBe("Null: null");
        expect(formRequest.clone().headers.get("Input")).toBe(
          "Input: string from a string input"
        );
        expect(formRequest.clone().headers.get("Textarea")).toBe(
          "Textarea: string from a textarea"
        );
        expect(formRequest.clone().headers.get("Checkbox-Checked")).toBe(
          "Checkbox (checked): true"
        );
        expect(formRequest.clone().headers.get("Checkbox-Unchecked")).toBe(
          "Checkbox (unchecked): false"
        );
        expect(formRequest.clone().headers.get("Input-Number")).toBe(
          "Input Number: 3"
        );

        expect(formRequest.clone().headers.get("Input-Floating-Number")).toBe(
          "Input Floating Number: 4.99"
        );
        expect(formRequest.clone().headers.get("Input-Date")).toBe(
          "Input Date: 2025-12-24"
        );
        expect(formRequest.clone().headers.get("Select")).toBe("Select: pro");
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
                  key: "String",
                  value: {
                    type: "string",
                    value: "String: {{query.user.data.status}}",
                  },
                },
                {
                  key: "Boolean",
                  value: {
                    type: "string",
                    value: "Boolean: {{query.user.data.emailVerified}}",
                  },
                },
                {
                  key: "Number",
                  value: {
                    type: "string",
                    value: "Number: {{query.user.data.accountBalance}}",
                  },
                },
                {
                  key: "Null",
                  value: {
                    type: "string",
                    value: "Null: {{query.user.data.lastLogin}}",
                  },
                },
                {
                  key: "Input",
                  value: {
                    type: "string",
                    value: "Input: {{this.form.string}}",
                  },
                },
                {
                  key: "Textarea",
                  value: {
                    type: "string",
                    value: "Textarea: {{this.form.textarea}}",
                  },
                },
                {
                  key: "Checkbox-Checked",
                  value: {
                    type: "string",
                    value: "Checkbox (checked): {{this.form.checkbox-checked}}",
                  },
                },
                {
                  key: "Checkbox-Unchecked",
                  value: {
                    type: "string",
                    value:
                      "Checkbox (unchecked): {{this.form.checkbox-unchecked}}",
                  },
                },
                {
                  key: "Input-Number",
                  value: {
                    type: "string",
                    value: "Input Number: {{this.form.number}}",
                  },
                },
                {
                  key: "Input-Floating-Number",
                  value: {
                    type: "string",
                    value:
                      "Input Floating Number: {{this.form.floating-number}}",
                  },
                },
                {
                  key: "Input-Date",
                  value: {
                    type: "string",
                    value: "Input Date: {{this.form.date}}",
                  },
                },
                {
                  key: "Select",
                  value: {
                    type: "string",
                    value: "Select: {{this.form.select}}",
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

        expect(text["String"]).toBe("String: ok");
        expect(text["Boolean"]).toBe("Boolean: true");
        expect(text["Number"]).toBe("Number: 19.99");
        expect(text["Null"]).toBe("Null: null");
        expect(text["Input"]).toBe("Input: string from a string input");
        expect(text["Textarea"]).toBe("Textarea: string from a textarea");
        expect(text["Checkbox-Checked"]).toBe("Checkbox (checked): true");
        expect(text["Checkbox-Unchecked"]).toBe("Checkbox (unchecked): false");
        expect(text["Input-Number"]).toBe("Input Number: 3");
        expect(text["Input-Floating-Number"]).toBe(
          "Input Floating Number: 4.99"
        );
        expect(text["Input-Date"]).toBe("Input Date: 2025-12-24");
        expect(text["Select"]).toBe("Select: pro");
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

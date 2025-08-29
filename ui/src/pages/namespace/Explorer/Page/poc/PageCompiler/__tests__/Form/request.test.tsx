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
    values: { type: "array", value: ["free", "pro", "enterprise"] },
    defaultValue: "pro",
  },
];

describe("form request", () => {
  describe("url", () => {
    test("interpolates variables in URL path", async () => {
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
    test("interpolates variables in query parameters", async () => {
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
                  key: "input-string",
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
          "?string=ok&boolean=true&number=19.99&null=null&input-string=string+from+a+string+input&textarea=string+from+a+textarea&checkbox-checked=true&checkbox-unchecked=false&input-number=3&input-floating-number=4.99&input-date=2025-12-24&select=pro"
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
    test("interpolates variables in request headers", async () => {
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
                  key: "Input-String",
                  value: "Input String: {{this.form.string}}",
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
        expect(formRequest.clone().headers.get("Input-String")).toBe(
          "Input String: string from a string input"
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
    test("interpolates variables in request body", async () => {
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
                  key: "string",
                  value: {
                    type: "string",
                    value: "String: {{query.user.data.status}}",
                  },
                },
                {
                  key: "boolean",
                  value: {
                    type: "string",
                    value: "Boolean: {{query.user.data.emailVerified}}",
                  },
                },
                {
                  key: "number",
                  value: {
                    type: "string",
                    value: "Number: {{query.user.data.accountBalance}}",
                  },
                },
                {
                  key: "null",
                  value: {
                    type: "string",
                    value: "Null: {{query.user.data.lastLogin}}",
                  },
                },
                {
                  key: "input-string",
                  value: {
                    type: "string",
                    value: "Input String: {{this.form.string}}",
                  },
                },
                {
                  key: "textarea",
                  value: {
                    type: "string",
                    value: "Textarea: {{this.form.textarea}}",
                  },
                },
                {
                  key: "checkbox-checked",
                  value: {
                    type: "string",
                    value: "Checkbox (checked): {{this.form.checkbox-checked}}",
                  },
                },
                {
                  key: "checkbox-unchecked",
                  value: {
                    type: "string",
                    value:
                      "Checkbox (unchecked): {{this.form.checkbox-unchecked}}",
                  },
                },
                {
                  key: "input-number",
                  value: {
                    type: "string",
                    value: "Input Number: {{this.form.number}}",
                  },
                },
                {
                  key: "input-floating-number",
                  value: {
                    type: "string",
                    value:
                      "Input Floating Number: {{this.form.floating-number}}",
                  },
                },
                {
                  key: "input-date",
                  value: {
                    type: "string",
                    value: "Input Date: {{this.form.date}}",
                  },
                },
                {
                  key: "select",
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
        const jsonResponse = JSON.parse(await formRequest.clone().text());

        expect(jsonResponse["string"]).toBe("String: ok");
        expect(jsonResponse["boolean"]).toBe("Boolean: true");
        expect(jsonResponse["number"]).toBe("Number: 19.99");
        expect(jsonResponse["null"]).toBe("Null: null");
        expect(jsonResponse["input-string"]).toBe(
          "Input String: string from a string input"
        );
        expect(jsonResponse["textarea"]).toBe(
          "Textarea: string from a textarea"
        );
        expect(jsonResponse["checkbox-checked"]).toBe(
          "Checkbox (checked): true"
        );
        expect(jsonResponse["checkbox-unchecked"]).toBe(
          "Checkbox (unchecked): false"
        );
        expect(jsonResponse["input-number"]).toBe("Input Number: 3");
        expect(jsonResponse["input-floating-number"]).toBe(
          "Input Floating Number: 4.99"
        );
        expect(jsonResponse["input-date"]).toBe("Input Date: 2025-12-24");
        expect(jsonResponse["select"]).toBe("Select: pro");
      });
    });

    test("resolves variable pointer in request body", async () => {
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
                  key: "string",
                  value: {
                    type: "variable",
                    value: "query.user.data.status",
                  },
                },
                {
                  key: "boolean",
                  value: {
                    type: "variable",
                    value: "query.user.data.emailVerified",
                  },
                },
                {
                  key: "number",
                  value: {
                    type: "variable",
                    value: "query.user.data.accountBalance",
                  },
                },
                {
                  key: "null",
                  value: {
                    type: "variable",
                    value: "query.user.data.lastLogin",
                  },
                },
                {
                  key: "input-string",
                  value: {
                    type: "variable",
                    value: "this.form.string",
                  },
                },
                {
                  key: "textarea",
                  value: {
                    type: "variable",
                    value: "this.form.textarea",
                  },
                },
                {
                  key: "checkbox-checked",
                  value: {
                    type: "variable",
                    value: "this.form.checkbox-checked",
                  },
                },
                {
                  key: "checkbox-unchecked",
                  value: {
                    type: "variable",
                    value: "this.form.checkbox-unchecked",
                  },
                },
                {
                  key: "input-number",
                  value: {
                    type: "variable",
                    value: "this.form.number",
                  },
                },
                {
                  key: "input-floating-number",
                  value: {
                    type: "variable",
                    value: "this.form.floating-number",
                  },
                },
                {
                  key: "input-date",
                  value: {
                    type: "variable",
                    value: "this.form.date",
                  },
                },
                {
                  key: "select",
                  value: {
                    type: "variable",
                    value: "this.form.select",
                  },
                },
                {
                  key: "array",
                  value: {
                    type: "variable",
                    value: "query.user.meta.subscriptionPlanOptions",
                  },
                },
                {
                  key: "object",
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
        const jsonResponse = JSON.parse(await formRequest.clone().text());

        expect(jsonResponse["string"]).toBe("ok");
        expect(jsonResponse["boolean"]).toBe(true);
        expect(jsonResponse["number"]).toBe(19.99);
        expect(jsonResponse["null"]).toBe(null);
        expect(jsonResponse["input-string"]).toBe("string from a string input");
        expect(jsonResponse["textarea"]).toBe("string from a textarea");
        expect(jsonResponse["checkbox-checked"]).toBe(true);
        expect(jsonResponse["checkbox-unchecked"]).toBe(false);
        expect(jsonResponse["input-number"]).toBe(3);
        expect(jsonResponse["input-floating-number"]).toBe(4.99);
        expect(jsonResponse["input-date"]).toBe("2025-12-24");
        expect(jsonResponse["select"]).toBe("pro");
        expect(jsonResponse["array"]).toEqual([
          { label: "Free Plan", value: "free" },
          { label: "Pro Plan", value: "pro" },
          { label: "Enterprise Plan", value: "enterprise" },
        ]);
        expect(jsonResponse["object"]).toEqual({
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

    test("can use booleans in request body", async () => {
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
                  key: "boolean-true",
                  value: {
                    type: "boolean",
                    value: true,
                  },
                },
                {
                  key: "boolean-false",
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
        const jsonResponse = JSON.parse(await formRequest.clone().text());

        expect(jsonResponse["boolean-true"]).toBe(true);
        expect(jsonResponse["boolean-false"]).toBe(false);
      });
    });

    test("can use numbers in request body", async () => {
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
                  key: "number-integer",
                  value: {
                    type: "number",
                    value: 3,
                  },
                },
                {
                  key: "number-float",
                  value: {
                    type: "number",
                    value: 4.99,
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
        const jsonResponse = JSON.parse(await formRequest.clone().text());

        expect(jsonResponse["number-integer"]).toBe(3);
        expect(jsonResponse["number-float"]).toBe(4.99);
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

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
    values: { type: "array", value: ["free", "pro", "enterprise"] },
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
        const requestUrl = new URL(formRequest.url);
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
        const requestUrl = new URL(formRequest.url);
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
        expect(formRequest.headers.get("String-Value")).toBe("String: ok");
        expect(formRequest.headers.get("Boolean-Value")).toBe("Boolean: true");
        expect(formRequest.headers.get("Number-Value")).toBe("Number: 19.99");
        expect(formRequest.headers.get("Null-Value")).toBe("Null: null");
        expect(formRequest.headers.get("Form-String-Value")).toBe(
          "String: string from a string input"
        );
        expect(formRequest.headers.get("Form-Textarea-Value")).toBe(
          "Textarea: string from a textarea"
        );
        expect(formRequest.headers.get("Form-Checkbox-Checked-Value")).toBe(
          "Checkbox: true"
        );
        expect(formRequest.headers.get("Form-Checkbox-Unchecked-Value")).toBe(
          "Checkbox: false"
        );
        expect(formRequest.headers.get("Form-Number-Value")).toBe("Number: 3");
        expect(formRequest.headers.get("Form-Date-Value")).toBe(
          "Date: 2025-12-24"
        );
        expect(formRequest.headers.get("Form-Select-Value")).toBe(
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
    test("resolves all variable types in request body", async () => {
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

      await waitFor(async () => {
        expect(apiRequestMock).toHaveBeenCalledTimes(1);
        const formRequest = apiRequestMock.mock.calls[0][0].request as Request;
        const text = JSON.parse(await formRequest.text());

        expect(text["String-Value"]).toBe("String: ok");
        expect(text["Boolean-Value"]).toBe("Boolean: true");
        expect(text["Number-Value"]).toBe("Number: 19.99");
        expect(text["Null-Value"]).toBe("Null: null");
        expect(text["Form-String-Value"]).toBe(
          "String: string from a string input"
        );
        expect(text["Form-Textarea-Value"]).toBe(
          "Textarea: string from a textarea"
        );
        expect(text["Form-Checkbox-Checked-Value"]).toBe("Checkbox: true");
        expect(text["Form-Checkbox-Unchecked-Value"]).toBe("Checkbox: false");
        expect(text["Form-Number-Value"]).toBe("Number: 3");
        expect(text["Form-Date-Value"]).toBe("Date: 2025-12-24");
        expect(text["Form-Select-Value"]).toBe("Select: pro");
      });
    });

    test("shows error for non-stringifiable variables in body", async () => {
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
});

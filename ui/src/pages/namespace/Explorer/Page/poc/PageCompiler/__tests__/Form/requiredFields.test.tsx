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
import userEvent from "@testing-library/user-event";

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

    test("submits the form when all required string input fields are filled in", async () => {
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
                  defaultValue: "default value",
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

  describe("number input field", () => {
    test("shows an error when a required number input field is missing a value", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm(
              [
                {
                  id: "number",
                  label: "number input",
                  description: "",
                  optional: false,
                  type: "form-number-input",
                  defaultValue: {
                    type: "number",
                    value: 0,
                  },
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

      // users clear the input field
      const user = userEvent.setup();
      await user.clear(
        screen.getByRole("spinbutton", { name: "number input" })
      );

      await screen.getByRole("button", { name: "save" }).click();
      expect(screen.getAllByText("Some required fields are missing (number)"));
    });

    test("submits the form when all required number input fields are filled in", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm(
              [
                {
                  id: "number",
                  label: "number input",
                  description: "",
                  optional: false,
                  type: "form-number-input",
                  defaultValue: {
                    type: "number",
                    value: 0,
                  },
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

  describe("date input field", () => {
    test("shows an error when a required date input field is missing a value", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm(
              [
                {
                  id: "date",
                  label: "date input",
                  description: "",
                  optional: false,
                  type: "form-date-input",
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
      expect(screen.getAllByText("Some required fields are missing (date)"));
    });

    test("submits the form when all required date input fields are filled in", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm(
              [
                {
                  id: "date",
                  label: "date input",
                  description: "",
                  optional: false,
                  type: "form-date-input",
                  defaultValue: "2025-12-24T00:00:00.000Z",
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

  describe("select field", () => {
    test("shows an error when a required select field is missing a value", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm(
              [
                {
                  id: "select",
                  label: "select",
                  description: "",
                  optional: false,
                  type: "form-select",
                  values: {
                    type: "array",
                    value: ["one", "two", "three"],
                  },
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
      expect(screen.getAllByText("Some required fields are missing (select)"));
    });

    test("submits the form when all required select fields are filled in", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm(
              [
                {
                  id: "select",
                  label: "select",
                  description: "",
                  optional: false,
                  type: "form-select",
                  values: {
                    type: "array",
                    value: ["one", "two", "three"],
                  },
                  defaultValue: "one",
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

  describe("checkbox field", () => {
    test("shows an error when a required checkbox field is missing a value", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm(
              [
                {
                  id: "checkbox",
                  label: "checkbox",
                  description: "my checkbox",
                  optional: false,
                  type: "form-checkbox",
                  defaultValue: {
                    type: "boolean",
                    value: false,
                  },
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
      expect(
        screen.getAllByText("Some required fields are missing (checkbox)")
      );
    });

    test("submits the form when all required checkbox fields are filled in", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm(
              [
                {
                  id: "checkbox",
                  label: "checkbox",
                  description: "my checkbox",
                  optional: false,
                  type: "form-checkbox",
                  defaultValue: {
                    type: "boolean",
                    value: true,
                  },
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

  describe("textarea field", () => {
    test("shows an error when a required textarea field is missing a value", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm(
              [
                {
                  id: "textarea",
                  label: "textarea",
                  description: "",
                  optional: false,
                  type: "form-textarea",
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
      expect(
        screen.getAllByText("Some required fields are missing (textarea)")
      );
    });

    test("submits the form when all required textarea fields are filled in", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm(
              [
                {
                  id: "textarea",
                  label: "textarea",
                  description: "",
                  optional: false,
                  type: "form-textarea",
                  defaultValue: "default value",
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

  describe("multiple fields", () => {
    test("shows an error when some required fields are missing and others are filled in", async () => {
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
                {
                  id: "number",
                  label: "number input",
                  description: "",
                  optional: false,
                  type: "form-number-input",
                  defaultValue: {
                    type: "number",
                    value: 0,
                  },
                },
                {
                  id: "select",
                  label: "select",
                  description: "",
                  optional: false,
                  type: "form-select",
                  values: {
                    type: "array",
                    value: ["one", "two", "three"],
                  },
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
      expect(
        screen.getAllByText("Some required fields are missing (string, select)")
      );
    });

    test("submits the form when all required fields of different types are filled in", async () => {
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
                  defaultValue: "test string",
                },
                {
                  id: "number",
                  label: "number input",
                  description: "",
                  optional: false,
                  type: "form-number-input",
                  defaultValue: {
                    type: "number",
                    value: 42,
                  },
                },
                {
                  id: "date",
                  label: "date input",
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
                  values: {
                    type: "array",
                    value: ["one", "two", "three"],
                  },
                  defaultValue: "one",
                },
                {
                  id: "checkbox",
                  label: "checkbox",
                  description: "my checkbox",
                  optional: false,
                  type: "form-checkbox",
                  defaultValue: {
                    type: "boolean",
                    value: true,
                  },
                },
                {
                  id: "textarea",
                  label: "textarea",
                  description: "",
                  optional: false,
                  type: "form-textarea",
                  defaultValue: "test textarea",
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

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

    test("submits the form when the required string input field is filled in", async () => {
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

    test("submits the form when the required number input field is filled in", async () => {
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

    test("submits the form when the required date input field is filled in", async () => {
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

    test("submits the form when the required select field is filled in", async () => {
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

    test("submits the form when the required checkbox field is filled in", async () => {
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

    test("submits the form when the required textarea field is filled in", async () => {
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

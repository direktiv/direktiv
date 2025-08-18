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

describe("default form values", () => {
  describe("valid", () => {
    test("string input can use string templates in the default value attribute", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm([
              {
                id: "string",
                label: "string input",
                description: "",
                optional: false,
                type: "form-string-input",
                variant: "text",
                defaultValue:
                  "a string input can use variable placeholders like string: {{query.user.data.status}}, number: {{query.user.data.userId}} and boolean: {{query.user.data.emailVerified}}",
              },
            ])}
            mode="live"
          />
        );
      });

      expect(
        (
          screen.getByRole("textbox", {
            name: "string input",
          }) as HTMLInputElement
        )?.value
      ).toBe(
        "a string input can use variable placeholders like string: ok, number: 101 and boolean: true"
      );
    });

    test("textarea can use string templates in the default value attribute", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm([
              {
                id: "textarea",
                label: "textarea",
                description: "",
                optional: false,
                type: "form-textarea",
                defaultValue:
                  "a textarea can use variable placeholders like string: {{query.user.data.status}}, number: {{query.user.data.userId}} and boolean: {{query.user.data.emailVerified}}",
              },
            ])}
            mode="live"
          />
        );
      });

      expect(
        (
          screen.getByRole("textbox", {
            name: "textarea",
          }) as HTMLInputElement
        )?.value
      ).toBe(
        "a textarea can use variable placeholders like string: ok, number: 101 and boolean: true"
      );
    });

    test("checkbox can be checked by default", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm([
              {
                id: "static-checkbox",
                label: "static checkbox",
                description: "default values is always checked",
                optional: false,
                type: "form-checkbox",
                defaultValue: {
                  type: "boolean",
                  value: true,
                },
              },
            ])}
            mode="live"
          />
        );
      });
      expect(
        screen.getByRole("checkbox", { name: "static checkbox", checked: true })
      );
    });

    test("checkbox can have a default value from a variable", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm([
              {
                id: "dynamic-checkbox",
                label: "dynamic checkbox",
                description:
                  "default value comes from the api ({{query.user.data.emailVerified}})",
                optional: false,
                type: "form-checkbox",
                defaultValue: {
                  type: "variable",
                  value: "query.user.data.emailVerified",
                },
              },
            ])}
            mode="live"
          />
        );
      });
      expect(
        screen.getByRole("checkbox", {
          name: "dynamic checkbox",
          checked: true,
        })
      );
    });
    test("number input can have a static default value", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm([
              {
                id: "static-number-input",
                label: "static number input",
                description: "default value is always 3",
                optional: false,
                type: "form-number-input",
                defaultValue: {
                  type: "number",
                  value: 3,
                },
              },
            ])}
            mode="live"
          />
        );
      });
      expect(
        (
          screen.getByRole("spinbutton", {
            name: "static number input",
          }) as HTMLInputElement
        )?.value
      ).toBe("3");
    });

    test("number input can have a default value from a variable", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm([
              {
                id: "dynamic-number-input",
                label: "dynamic number input",
                description:
                  "default value comes from API ({{query.user.data.accountBalance}})",
                optional: false,
                type: "form-number-input",
                defaultValue: {
                  type: "variable",
                  value: "query.user.data.accountBalance",
                },
              },
            ])}
            mode="live"
          />
        );
      });
      expect(
        (
          screen.getByRole("spinbutton", {
            name: "dynamic number input",
          }) as HTMLInputElement
        )?.value
      ).toBe("19.99");
    });

    test("date input can have a static default value", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm([
              {
                id: "static-date",
                label: "static date",
                description: "default value is always 2025-12-24",
                optional: false,
                type: "form-date-input",
                defaultValue: "2025-12-24T00:00:00.000Z",
              },
            ])}
            mode="live"
          />
        );
      });
      expect(screen.getByRole("button", { name: "December 24, 2025" }));
    });

    test("date input can have a default value from a variable", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm([
              {
                id: "dynamic-date",
                label: "dynamic date",
                description:
                  "default value comes from the api ({{query.user.data.membershipStartDate}})",
                optional: false,
                type: "form-date-input",
                defaultValue: "{{query.user.data.membershipStartDate}}",
              },
            ])}
            mode="live"
          />
        );
      });
      expect(screen.getByRole("button", { name: "June 15, 2023" }));
    });

    test("select input can have a static default value", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm([
              {
                id: "static-select",
                label: "static select",
                description: "default value is always two",
                optional: false,
                type: "form-select",
                values: ["free", "pro", "enterprise"],
                defaultValue: "free",
              },
            ])}
            mode="live"
          />
        );
      });
      expect(
        (
          screen.getByRole("combobox", {
            name: "static select",
          }) as HTMLInputElement
        )?.value
      ).toBe("free");
    });

    test("select input can have a default value from a variable", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm([
              {
                id: "dynamic-select",
                label: "dynamic select",
                description:
                  "default value comes from API ({{query.user.data.subscriptionPlan}})",
                optional: false,
                type: "form-select",
                values: ["free", "pro", "enterprise"],
                defaultValue: "{{query.user.data.subscriptionPlan}}",
              },
            ])}
            mode="live"
          />
        );
      });
      expect(
        (
          screen.getByRole("combobox", {
            name: "dynamic select",
          }) as HTMLInputElement
        )?.value
      ).toBe("pro");
    });
  });

  describe("invalid", () => {
    test("shows an error when textarea default value is an object", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm([
              {
                id: "textarea-using-an-object",
                label: "invalid textarea",
                description:
                  "This textarea is using an object in the template string",
                optional: false,
                type: "form-textarea",
                defaultValue: "{{query.user.data}}",
              },
            ])}
            mode="live"
          />
        );
      });

      await screen
        .getByRole("button", {
          name: "There was an unexpected error",
        })
        .click();
      expect(
        screen.getByText(
          "Pointing to a value that can not be stringified. Make sure to point to either a String, Number, Boolean, or Null."
        )
      );
    });

    test("shows an error when checkbox default value is a string", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm([
              {
                id: "checkbox-pointing-to-string",
                label: "invalid checkbox",
                description:
                  "This checkbox is pointing to a string for the default value",
                optional: false,
                type: "form-checkbox",
                defaultValue: {
                  type: "variable",
                  value: "query.user.data.status",
                },
              },
            ])}
            mode="live"
          />
        );
      });

      await screen
        .getByRole("button", {
          name: "There was an unexpected error",
        })
        .click();
      expect(screen.getByText("Pointing to a value that is not a boolean."));
    });

    test("shows an error when number input default value is a string", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm([
              {
                id: "number-using-string",
                label: "invalid number inpuit",
                description:
                  "This number input is pointing to a string for the default value",
                optional: false,
                type: "form-number-input",
                defaultValue: {
                  type: "variable",
                  value: "query.user.data.status",
                },
              },
            ])}
            mode="live"
          />
        );
      });

      await screen
        .getByRole("button", {
          name: "There was an unexpected error",
        })
        .click();
      expect(screen.getByText("Pointing to a value that is not a number."));
    });

    test("shows an error when select input default value is an object", async () => {
      await act(async () => {
        render(
          <PageCompiler
            setPage={setPage}
            page={createDirektivPageWithForm([
              {
                id: "select-pointing-to-object",
                label: "dynamic select",
                description:
                  "This select input is pointing to an object for the default value",
                optional: false,
                type: "form-select",
                values: ["free", "pro", "enterprise"],
                defaultValue: "{{query.user.data}}",
              },
            ])}
            mode="live"
          />
        );
      });

      await screen
        .getByRole("button", {
          name: "There was an unexpected error",
        })
        .click();
      expect(
        screen.getByText(
          "Pointing to a value that can not be stringified. Make sure to point to either a String, Number, Boolean, or Null."
        )
      );
    });
  });
});

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

describe("select input", () => {
  test("values can be sourced from a list of strings", async () => {
    await act(async () => {
      render(
        <PageCompiler
          setPage={setPage}
          page={createDirektivPageWithForm([
            {
              id: "dynamic-select",
              label: "dynamic select",
              description: "default value is always two",
              optional: false,
              type: "form-select",
              values: {
                type: "static-select-options",
                value: [
                  {
                    label: "One",
                    value: "one",
                  },
                  {
                    label: "Two",
                    value: "two",
                  },
                  {
                    label: "Three",
                    value: "three",
                  },
                ],
              },
              defaultValue: "",
            },
          ])}
          mode="live"
        />
      );
    });
    const select = screen.getByRole("combobox", {
      name: "dynamic select",
    }) as HTMLSelectElement;
    expect(select.parentElement?.textContent).toContain("OneTwoThree");
  });

  test("values can be sourced from a variable", async () => {
    await act(async () => {
      render(
        <PageCompiler
          setPage={setPage}
          page={createDirektivPageWithForm([
            {
              id: "dynamic-select",
              label: "dynamic select",
              description: "default value is always two",
              optional: false,
              type: "form-select",
              values: {
                type: "variable-select-options",
                data: "query.user.meta.subscriptionPlanOptions",
                label: "label",
                value: "value",
              },
              defaultValue: "",
            },
          ])}
          mode="live"
        />
      );
    });
    const select = screen.getByRole("combobox", {
      name: "dynamic select",
    }) as HTMLSelectElement;
    expect(select.parentElement?.textContent).contain("FreeProEnterprise");
  });

  test("shows error when using a non array variable", async () => {
    await act(async () => {
      render(
        <PageCompiler
          setPage={setPage}
          page={createDirektivPageWithForm([
            {
              id: "dynamic-select",
              label: "dynamic select",
              description: "default value is always two",
              optional: false,
              type: "form-select",
              values: {
                type: "variable-select-options",
                data: "query.user.meta",
                label: "label",
                value: "value",
              },
              defaultValue: "",
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
        "Variable error (query.user.meta): Pointing to a value that is not an array."
      )
    );
  });

  test("shows error when the path to the label does not exist", async () => {
    await act(async () => {
      render(
        <PageCompiler
          setPage={setPage}
          page={createDirektivPageWithForm([
            {
              id: "dynamic-select",
              label: "dynamic select",
              description: "default value is always two",
              optional: false,
              type: "form-select",
              values: {
                type: "variable-select-options",
                data: "query.user.meta.subscriptionPlanOptions",
                label: "this.label.does.not.exist",
                value: "value",
              },
              defaultValue: "",
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
        "Variable error (this.label.does.not.exist): The path is not valid. It points to an undefined value."
      )
    );
  });

  test("shows error when the path to the value does not exist", async () => {
    await act(async () => {
      render(
        <PageCompiler
          setPage={setPage}
          page={createDirektivPageWithForm([
            {
              id: "dynamic-select",
              label: "dynamic select",
              description: "default value is always two",
              optional: false,
              type: "form-select",
              values: {
                type: "variable-select-options",
                data: "query.user.meta.subscriptionPlanOptions",
                label: "label",
                value: "this.value.does.not.exist",
              },
              defaultValue: "",
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
        "Variable error (this.value.does.not.exist): The path is not valid. It points to an undefined value."
      )
    );
  });
});

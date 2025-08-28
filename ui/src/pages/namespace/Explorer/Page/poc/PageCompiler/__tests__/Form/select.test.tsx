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
                type: "string-array",
                value: ["one", "two", "three"],
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
    expect(select.parentElement?.textContent).toContain("onetwothree");
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
                type: "variable",
                value: "query.user.meta.subscriptionPlanOptions",
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
    expect(select.parentElement?.textContent).toContain("freeproenterprise");
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
                type: "variable",
                value: "query.user.meta",
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

  test("shows error when using an array but the values are not strings", async () => {
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
                type: "variable",
                value: "query.user.data.recentActivity",
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
        "Variable error (query.user.data.recentActivity): Pointing to a value that is not an array of strings."
      )
    );
  });
});

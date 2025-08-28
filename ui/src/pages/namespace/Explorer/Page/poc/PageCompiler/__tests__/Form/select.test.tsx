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
});

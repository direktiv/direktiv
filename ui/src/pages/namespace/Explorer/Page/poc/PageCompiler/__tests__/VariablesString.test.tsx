import { act, render, screen } from "@testing-library/react";
import { describe, expect, test, vi } from "vitest";

import { PageCompiler } from "..";
import { createPage } from "./utils";

vi.mock("react-i18next", () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

describe("VariableString", () => {
  test("will show an error when the variable has an invalid namespace", async () => {
    await act(async () => {
      render(
        <PageCompiler
          page={createPage([
            {
              type: "headline",
              size: "h1",
              label: "template string without id: {{thisDoesNotExist}}",
            },
          ])}
          mode="live"
        />
      );
    });

    expect(screen.getByRole("heading", { level: 1 }).textContent).toBe(
      "template string without id: thisDoesNotExist (namespaceInvalid)"
    );
  });

  test("will show an error when the variable has no id", async () => {
    await act(async () => {
      render(
        <PageCompiler
          page={createPage([
            {
              type: "headline",
              size: "h1",
              label: "template string without id: {{query}}",
            },
          ])}
          mode="live"
        />
      );
    });

    expect(screen.getByRole("heading", { level: 1 }).textContent).toBe(
      "template string without id: query (idUndefined)"
    );
  });

  test("will show an error when the variable has no pointer", async () => {
    await act(async () => {
      render(
        <PageCompiler
          page={createPage([
            {
              type: "headline",
              size: "h1",
              label: "template string without id: {{query.id}}",
            },
          ])}
          mode="live"
        />
      );
    });

    expect(screen.getByRole("heading", { level: 1 }).textContent).toBe(
      "template string without id: query.id (pointerUndefined)"
    );
  });

  test("will show an error when the variable will point to an undefined value", async () => {
    await act(async () => {
      render(
        <PageCompiler
          page={createPage([
            {
              type: "headline",
              size: "h1",
              label: "template string without id: {{query.id.nothing}}",
            },
          ])}
          mode="live"
        />
      );
    });

    expect(screen.getByRole("heading", { level: 1 }).textContent).toBe(
      "template string without id: query.id.nothing (NoStateForId)"
    );
  });
});

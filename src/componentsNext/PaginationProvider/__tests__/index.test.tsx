import { describe, expect, test } from "vitest";
import { render, screen } from "@testing-library/react";

import PaginationProvider from "..";

const items = [
  { id: 1, name: "Item 1" },
  { id: 2, name: "Item 2" },
  { id: 3, name: "Item 3" },
  { id: 4, name: "Item 4" },
  { id: 5, name: "Item 5" },
  { id: 6, name: "Item 6" },
  { id: 7, name: "Item 7" },
  { id: 8, name: "Item 8" },
  { id: 9, name: "Item 9" },
  { id: 10, name: "Item 10" },
  { id: 11, name: "Item 11" },
  { id: 12, name: "Item 12" },
  { id: 13, name: "Item 13" },
];

describe("Pagination Provider", () => {
  test("should render with proper pagination logic", () => {
    render(
      <PaginationProvider items={items}>
        {({
          currentItems,
          gotoFirstPage,
          gotoLastPage,
          gotoNextPage,
          gotoPreviousPage,
          isFirstPage,
          isLastPage,
          page,
        }) => (
          <div>
            <ul>
              {currentItems.map((item) => (
                <li key={item.id}>{item.name}</li>
              ))}
            </ul>
            <h1 data-testid="isFirstPage">
              Is first Page? {isFirstPage ? "yes" : "no"}
            </h1>
            <h1 data-testid="isLastPage">
              Is last Page? {isLastPage ? "yes" : "no"}
            </h1>
            <button data-testid="gotoFirstPage" onClick={gotoFirstPage}>
              go to first page
            </button>
            <button data-testid="gotoPreviousPage" onClick={gotoPreviousPage}>
              go to previous Page
            </button>
            <span data-testid="page">{page}</span>
            <button data-testid="gotoNextPage" onClick={gotoNextPage}>
              go to next page
            </button>
            <button data-testid="gotoLastPage" onClick={gotoLastPage}>
              go to last page
            </button>
          </div>
        )}
      </PaginationProvider>
    );

    expect(screen.getByTestId("isFirstPage").textContent).toContain("yes");
    expect(screen.getByTestId("isLastPage").textContent).toContain("no");
  });
});

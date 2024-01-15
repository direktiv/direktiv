import { act, render, screen } from "@testing-library/react";
import { describe, expect, test } from "vitest";

import PaginationProvider from "..";
import userEvent from "@testing-library/user-event";

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

const user = userEvent.setup();

describe("Pagination Provider", () => {
  test("should render with proper pagination logic", async () => {
    render(
      <PaginationProvider items={items} pageSize={4}>
        {({
          currentItems,
          goToFirstPage,
          goToLastPage,
          goToNextPage,
          goToPreviousPage,
          isFirstPage,
          isLastPage,
          currentPage,
          pagesList,
          totalPages,
        }) => (
          <div>
            <ul data-testid="itemsList">
              {currentItems.map((item) => (
                <li key={item.id}>{item.id}</li>
              ))}
            </ul>
            <h1 data-testid="isFirstPage">
              Is first Page? {isFirstPage ? "yes" : "no"}
            </h1>
            <h1 data-testid="pages">{totalPages} pages</h1>
            <h1 data-testid="pagesArray">{pagesList.join(",")}</h1>
            <h1 data-testid="isLastPage">
              Is last Page? {isLastPage ? "yes" : "no"}
            </h1>
            <button data-testid="goToFirstPage" onClick={goToFirstPage}>
              go to first page
            </button>
            <button data-testid="goToPreviousPage" onClick={goToPreviousPage}>
              go to previous Page
            </button>
            <span data-testid="page">{currentPage}</span>
            <button data-testid="goToNextPage" onClick={goToNextPage}>
              go to next page
            </button>
            <button data-testid="goToLastPage" onClick={goToLastPage}>
              go to last page
            </button>
          </div>
        )}
      </PaginationProvider>
    );

    const nextPageButton = screen.getByTestId("goToNextPage");
    const prevPageButton = screen.getByTestId("goToPreviousPage");
    const firstPageButton = screen.getByTestId("goToFirstPage");
    const lastPageButton = screen.getByTestId("goToLastPage");

    // page one by default
    expect(screen.getByTestId("isFirstPage").textContent).toContain("yes");
    expect(screen.getByTestId("isLastPage").textContent).toContain("no");
    expect(screen.getByTestId("page").textContent).toBe("1");
    expect(screen.getByTestId("pages").textContent).toBe("4 pages");
    expect(screen.getByTestId("pagesArray").textContent).toBe("1,2,3,4");
    expect(screen.getByTestId("itemsList").textContent).toBe("1234");

    // clicking on previous page should not change anything
    await act(async () => {
      await user.click(prevPageButton);
    });

    expect(screen.getByTestId("isFirstPage").textContent).toContain("yes");
    expect(screen.getByTestId("isLastPage").textContent).toContain("no");
    expect(screen.getByTestId("page").textContent).toBe("1");
    expect(screen.getByTestId("pages").textContent).toBe("4 pages");
    expect(screen.getByTestId("pagesArray").textContent).toBe("1,2,3,4");
    expect(screen.getByTestId("itemsList").textContent).toBe("1234");

    // clicking on first page should not change anything
    await act(async () => {
      await user.click(firstPageButton);
    });

    expect(screen.getByTestId("isFirstPage").textContent).toContain("yes");
    expect(screen.getByTestId("isLastPage").textContent).toContain("no");
    expect(screen.getByTestId("page").textContent).toBe("1");
    expect(screen.getByTestId("pages").textContent).toBe("4 pages");
    expect(screen.getByTestId("pagesArray").textContent).toBe("1,2,3,4");
    expect(screen.getByTestId("itemsList").textContent).toBe("1234");

    // page 2
    await act(async () => {
      await user.click(nextPageButton);
    });

    expect(screen.getByTestId("isFirstPage").textContent).toContain("no");
    expect(screen.getByTestId("isLastPage").textContent).toContain("no");
    expect(screen.getByTestId("page").textContent).toBe("2");
    expect(screen.getByTestId("pages").textContent).toBe("4 pages");
    expect(screen.getByTestId("pagesArray").textContent).toBe("1,2,3,4");
    expect(screen.getByTestId("itemsList").textContent).toBe("5678");

    // page 3
    await act(async () => {
      await user.click(nextPageButton);
    });

    expect(screen.getByTestId("isFirstPage").textContent).toContain("no");
    expect(screen.getByTestId("isLastPage").textContent).toContain("no");
    expect(screen.getByTestId("page").textContent).toBe("3");
    expect(screen.getByTestId("pages").textContent).toBe("4 pages");
    expect(screen.getByTestId("pagesArray").textContent).toBe("1,2,3,4");
    expect(screen.getByTestId("itemsList").textContent).toBe("9101112");

    // page 4 (last page)
    await act(async () => {
      await user.click(nextPageButton);
    });

    expect(screen.getByTestId("isFirstPage").textContent).toContain("no");
    expect(screen.getByTestId("isLastPage").textContent).toContain("yes");
    expect(screen.getByTestId("page").textContent).toBe("4");
    expect(screen.getByTestId("pages").textContent).toBe("4 pages");
    expect(screen.getByTestId("itemsList").textContent).toBe("13");

    // clicking on next page again should not change anything
    await act(async () => {
      await user.click(nextPageButton);
    });

    expect(screen.getByTestId("isFirstPage").textContent).toContain("no");
    expect(screen.getByTestId("isLastPage").textContent).toContain("yes");
    expect(screen.getByTestId("page").textContent).toBe("4");
    expect(screen.getByTestId("pages").textContent).toBe("4 pages");
    expect(screen.getByTestId("pagesArray").textContent).toBe("1,2,3,4");
    expect(screen.getByTestId("itemsList").textContent).toBe("13");

    // go to last page should not change anything
    await act(async () => {
      await user.click(lastPageButton);
    });

    expect(screen.getByTestId("isFirstPage").textContent).toContain("no");
    expect(screen.getByTestId("isLastPage").textContent).toContain("yes");
    expect(screen.getByTestId("page").textContent).toBe("4");
    expect(screen.getByTestId("pages").textContent).toBe("4 pages");
    expect(screen.getByTestId("pagesArray").textContent).toBe("1,2,3,4");
    expect(screen.getByTestId("itemsList").textContent).toBe("13");

    // go to first page
    await act(async () => {
      await user.click(firstPageButton);
    });

    expect(screen.getByTestId("isFirstPage").textContent).toContain("yes");
    expect(screen.getByTestId("isLastPage").textContent).toContain("no");
    expect(screen.getByTestId("page").textContent).toBe("1");
    expect(screen.getByTestId("pages").textContent).toBe("4 pages");
    expect(screen.getByTestId("pagesArray").textContent).toBe("1,2,3,4");
    expect(screen.getByTestId("itemsList").textContent).toBe("1234");

    // go to last page
    await act(async () => {
      await user.click(lastPageButton);
    });

    expect(screen.getByTestId("isFirstPage").textContent).toContain("no");
    expect(screen.getByTestId("isLastPage").textContent).toContain("yes");
    expect(screen.getByTestId("page").textContent).toBe("4");
    expect(screen.getByTestId("pages").textContent).toBe("4 pages");
    expect(screen.getByTestId("pagesArray").textContent).toBe("1,2,3,4");
    expect(screen.getByTestId("itemsList").textContent).toBe("13");
  });
});

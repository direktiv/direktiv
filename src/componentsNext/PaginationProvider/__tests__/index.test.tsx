import { describe, expect, test } from "vitest";

import PaginationProvider from "..";
import { render } from "@testing-library/react";

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
  test("should render", () => {
    const { debug } = render(
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
            <h1>Is first Page? {isFirstPage ? "yes" : "no"}</h1>
            <h1>Is last Page? {isLastPage ? "yes" : "no"}</h1>
            <button onClick={gotoFirstPage}>go to first page</button>
            <button onClick={gotoPreviousPage}>go to previous Page</button>
            {page}
            <button onClick={gotoNextPage}>go to next page</button>
            <button onClick={gotoLastPage}>go to last page</button>
          </div>
        )}
      </PaginationProvider>
    );
  });
});

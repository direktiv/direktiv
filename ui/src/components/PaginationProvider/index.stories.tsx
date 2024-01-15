import { Pagination, PaginationLink } from "~/design/Pagination";
import React, { useState } from "react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import PaginationProvider from ".";

export default {
  title: "Components/Pagination Provider",
};

type Person = {
  id: number;
  name: string;
};

const persons: Person[] = [
  { id: 1, name: "Person 1" },
  { id: 2, name: "Person 2" },
  { id: 3, name: "Person 3" },
  { id: 4, name: "Person 4" },
  { id: 5, name: "Person 5" },
  { id: 6, name: "Person 6" },
  { id: 7, name: "Person 7" },
  { id: 8, name: "Person 8" },
  { id: 9, name: "Person 9" },
  { id: 10, name: "Person 10" },
  { id: 11, name: "Person 11" },
  { id: 12, name: "Person 12" },
  { id: 13, name: "Person 13" },
];

export const Default = () => {
  const [pageSize, setPageSize] = useState(3);
  return (
    <div
      className="flex flex-col items-center space-y-5 p-5
    "
    >
      <PaginationProvider items={persons} pageSize={pageSize}>
        {({
          currentItems,
          goToFirstPage,
          goToPage,
          goToNextPage,
          goToPreviousPage,
          isFirstPage,
          isLastPage,
          currentPage,
          totalPages,
          pagesList,
        }) => (
          <>
            <div className="grid gap-y-2 rounded border p-5 text-center shadow-sm">
              <div>is first page? {isFirstPage ? "👍" : "👎"}</div>
              <div>is last page? {isLastPage ? "👍" : "👎"}</div>
              <div>
                page {currentPage} / {totalPages}
              </div>
              <div className="flex items-center space-x-3">
                <span>page size</span>
                <Select
                  onValueChange={(value) => {
                    setPageSize(parseInt(value));
                    goToFirstPage();
                  }}
                >
                  <SelectTrigger>
                    <SelectValue placeholder={pageSize} />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="3">3</SelectItem>
                    <SelectItem value="4">4</SelectItem>
                    <SelectItem value="5">5</SelectItem>
                    <SelectItem value="6">6</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div className="grid gap-y-2 rounded border p-5 text-center shadow-sm">
              {currentItems.map((person) => (
                <div key={person.id}>{person.name}</div>
              ))}
            </div>
            <Pagination align="center">
              <PaginationLink icon="left" onClick={() => goToPreviousPage()} />
              {pagesList.map((p) => (
                <PaginationLink
                  active={currentPage === p}
                  key={`${p}`}
                  onClick={() => goToPage(p)}
                >
                  {p}
                </PaginationLink>
              ))}
              <PaginationLink icon="right" onClick={() => goToNextPage()} />
            </Pagination>
          </>
        )}
      </PaginationProvider>
    </div>
  );
};

import { Pagination, PaginationLink } from "../../design/Pagination";
import React, { useState } from "react";

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

export const Default = () => (
  <PaginationProvider items={persons} pageSize={3}>
    {({
      currentItems,
      gotoFirstPage,
      gotoLastPage,
      gotoNextPage,
      gotoPreviousPage,
      isFirstPage,
      isLastPage,
      page,
      pagesCount,
      gotoPage,
      pages,
    }) => (
      <>
        <div>
          {currentItems.map((person) => (
            <div key={person.id}>{person.name}</div>
          ))}
        </div>

        <Pagination align="center">
          <PaginationLink icon="left" onClick={() => gotoPreviousPage()} />
          {pages.map((p) => (
            <PaginationLink
              active={page === p}
              key={`${p}`}
              onClick={() => gotoPage(p)}
            >
              {p}
            </PaginationLink>
          ))}
          <PaginationLink icon="right" onClick={() => gotoNextPage()} />
        </Pagination>
      </>
    )}
  </PaginationProvider>
);

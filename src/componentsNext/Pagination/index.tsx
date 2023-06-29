import { Dispatch, SetStateAction } from "react";
import {
  PaginationLink,
  Pagination as PaginationWrapper,
} from "~/design/Pagination";

import describePagination from "./describePagination";

type SetState<T> = Dispatch<SetStateAction<T>>;

export const Pagination = ({
  itemsPerPage,
  totalItems,
  offset,
  setOffset,
}: {
  itemsPerPage: number;
  totalItems?: number;
  offset: number;
  setOffset: SetState<number>;
}) => {
  const setOffsetByPageNumber = (pageNumber: number) =>
    (pageNumber - 1) * itemsPerPage;

  const numberOfItems = totalItems ?? 0;
  const pages = Math.ceil(numberOfItems / itemsPerPage);
  const currentPage = Math.ceil(offset / itemsPerPage) + 1;
  const isFirstPage = currentPage === 1;
  const isLastPage = currentPage === pages;

  const previousPage = currentPage > 1 ? currentPage - 1 : null;
  const nextPage = currentPage < pages ? currentPage + 1 : null;

  const paginationDescription = describePagination({ currentPage, pages });

  return (
    <PaginationWrapper>
      <PaginationLink
        icon="left"
        onClick={() =>
          previousPage && setOffset(setOffsetByPageNumber(previousPage))
        }
        disabled={isFirstPage}
      />
      {paginationDescription.map((page) => {
        if (page === "...")
          return (
            <PaginationLink key={page} disabled>
              ...
            </PaginationLink>
          );

        const isActive = currentPage === page;
        return (
          <PaginationLink
            key={page}
            active={isActive}
            onClick={() => {
              !isActive && setOffset(setOffsetByPageNumber(page));
            }}
          >
            {page}
          </PaginationLink>
        );
      })}
      <PaginationLink
        icon="right"
        onClick={() => nextPage && setOffset(setOffsetByPageNumber(nextPage))}
        disabled={isLastPage}
      />
    </PaginationWrapper>
  );
};

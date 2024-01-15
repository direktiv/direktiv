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
        data-testid="pagination-btn-left"
      />
      {paginationDescription.map((page, index) => {
        const isActive = currentPage === page;
        const isEllipsis = page === "…";
        return (
          <PaginationLink
            key={index}
            active={isActive}
            onClick={() => {
              !isEllipsis &&
                !isActive &&
                setOffset(setOffsetByPageNumber(page));
            }}
            disabled={isEllipsis}
            data-testid={`pagination-btn-page-${page}`}
          >
            {page}
          </PaginationLink>
        );
      })}
      <PaginationLink
        icon="right"
        onClick={() => nextPage && setOffset(setOffsetByPageNumber(nextPage))}
        disabled={isLastPage}
        data-testid="pagination-btn-right"
      />
    </PaginationWrapper>
  );
};

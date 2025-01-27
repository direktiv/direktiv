import {
  PaginationLink,
  Pagination as PaginationWrapper,
} from "~/design/Pagination";

import describePagination from "./describePagination";

export const Pagination = ({
  itemsPerPage,
  totalItems,
  value,
  onChange,
}: {
  itemsPerPage: number;
  totalItems?: number;
  value: number;
  onChange: (page: number) => void;
}) => {
  const setOffsetByPageNumber = (pageNumber: number) =>
    (pageNumber - 1) * itemsPerPage;

  const numberOfItems = totalItems ?? 0;
  const pages = Math.max(1, Math.ceil(numberOfItems / itemsPerPage));
  const currentPage = Math.ceil(value / itemsPerPage) + 1;
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
          previousPage && onChange(setOffsetByPageNumber(previousPage))
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
              !isEllipsis && !isActive && onChange(setOffsetByPageNumber(page));
            }}
            disabled={(isFirstPage && isLastPage) || isEllipsis}
            data-testid={`pagination-btn-page-${page}`}
          >
            {page}
          </PaginationLink>
        );
      })}
      <PaginationLink
        icon="right"
        onClick={() => nextPage && onChange(setOffsetByPageNumber(nextPage))}
        disabled={isLastPage}
        data-testid="pagination-btn-right"
      />
    </PaginationWrapper>
  );
};

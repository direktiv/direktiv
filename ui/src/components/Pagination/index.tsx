import {
  PaginationLink,
  Pagination as PaginationWrapper,
} from "~/design/Pagination";

import describePagination from "./describePagination";

export const Pagination = ({
  totalPages,
  value,
  onChange,
}: {
  totalPages: number;
  value: number;
  onChange: (page: number) => void;
}) => {
  const isFirstPage = value === 1;

  const isLastPage = value === totalPages;

  const previousPage = value > 1 ? value - 1 : null;

  const nextPage = value < totalPages ? value + 1 : null;

  const paginationDescription = describePagination({
    currentPage: value,
    totalPages,
  });

  return (
    <PaginationWrapper>
      <PaginationLink
        icon="left"
        onClick={() => previousPage && onChange(previousPage)}
        disabled={isFirstPage}
        data-testid="pagination-btn-left"
      />
      {paginationDescription.map((page, index) => {
        const isActive = value === page;
        const isEllipsis = page === "…";
        return (
          <PaginationLink
            key={index}
            active={isActive}
            onClick={() => {
              !isEllipsis && !isActive && onChange(page);
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
        onClick={() => nextPage && onChange(nextPage)}
        disabled={isLastPage}
        data-testid="pagination-btn-right"
      />
    </PaginationWrapper>
  );
};

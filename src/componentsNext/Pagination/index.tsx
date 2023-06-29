import { Dispatch, SetStateAction } from "react";
import {
  PaginationLink,
  Pagination as PaginationWrapper,
} from "~/design/Pagination";

type SetState<T> = Dispatch<SetStateAction<T>>;

const Pagination = ({
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
  const numberOfInstances = totalItems ?? 0;
  const pages = Math.ceil(numberOfInstances / itemsPerPage);
  const currentPage = Math.ceil(offset / itemsPerPage) + 1;
  const isFirstPage = currentPage === 1;
  const isLastPage = currentPage === pages;

  return (
    <PaginationWrapper>
      <PaginationLink
        icon="left"
        onClick={() => setOffset(0)}
        disabled={isFirstPage}
      />
      {[...Array(pages)].map((_, i) => {
        const pageNumber = i + 1;
        const isActive = currentPage === pageNumber;
        return (
          <PaginationLink
            key={i}
            active={isActive}
            onClick={() => {
              isActive ? null : setOffset((pageNumber - 1) * itemsPerPage);
            }}
          >
            {pageNumber}
          </PaginationLink>
        );
      })}

      <PaginationLink
        icon="right"
        onClick={() => setOffset((pages - 1) * itemsPerPage)}
        disabled={isLastPage}
      />
    </PaginationWrapper>
  );
};

export default Pagination;

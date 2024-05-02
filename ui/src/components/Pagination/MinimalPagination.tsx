import { Dispatch, SetStateAction } from "react";
import {
  PaginationLink,
  Pagination as PaginationWrapper,
} from "~/design/Pagination";

type SetState<T> = Dispatch<SetStateAction<T>>;

type MinimalPaginationProps = {
  itemsPerPage: number;
  totalItems: number;
  offset: number;
  setOffset: SetState<number>;
  isLastPage: boolean;
};

export const MinimalPagination = ({
  itemsPerPage,
  totalItems,
  offset,
  setOffset,
  isLastPage,
}: MinimalPaginationProps) => {
  const isFirstPage = offset === 0;

  return (
    <PaginationWrapper>
      <PaginationLink
        icon="left"
        onClick={() => setOffset(offset - itemsPerPage)}
        disabled={isFirstPage}
        data-testid="pagination-btn-left"
      />

      <PaginationLink
        icon="right"
        onClick={() => setOffset(offset + itemsPerPage)}
        disabled={isLastPage}
        data-testid="pagination-btn-right"
      />
    </PaginationWrapper>
  );
};

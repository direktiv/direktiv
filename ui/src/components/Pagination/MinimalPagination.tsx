import { Dispatch, SetStateAction } from "react";
import {
  PaginationLink,
  Pagination as PaginationWrapper,
} from "~/design/Pagination";

type SetState<T> = Dispatch<SetStateAction<T>>;

export const MinimalPagination = ({
  itemsPerPage,
  offset,
  setOffset,
  isLastPage,
}: {
  itemsPerPage: number;
  offset: number;
  setOffset: SetState<number>;
  isLastPage: boolean;
}) => {
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

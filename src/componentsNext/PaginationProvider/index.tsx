import { useState } from "react";

const PaginationProvider = <TArrayItem,>({
  children,
  items,
  pageSize: pageSizeProp,
}: {
  children: (props: {
    currentItems: TArrayItem[];
    goToPage: (page: number) => void;
    goToFirstPage: () => void;
    goToLastPage: () => void;
    goToNextPage: () => void;
    goToPreviousPage: () => void;
    isFirstPage: boolean;
    isLastPage: boolean;
    page: number;
    pages: number[];
    pagesCount: number;
  }) => JSX.Element;
  pageSize?: number;
  items: TArrayItem[];
}) => {
  const firstPage = 1;
  const [page, setPage] = useState(firstPage);
  const pageSize = pageSizeProp || 10;
  const lastPage = Math.ceil(items.length / pageSize);
  const isLastPage = page === lastPage;
  const isFirstPage = page === firstPage;

  const sliceStart = (page - 1) * pageSize;
  const sliceEnd = sliceStart + pageSize;
  const currentItems = items.slice(sliceStart, sliceEnd);

  const goToFirstPage = () => setPage(1);
  const goToLastPage = () => setPage(lastPage);
  const goToNextPage = () =>
    setPage((page) => (page < lastPage ? page + 1 : page));
  const goToPreviousPage = () =>
    setPage((page) => (page > firstPage ? page - 1 : page));
  const goToPage = (page: number) => {
    if (page >= firstPage && page <= lastPage) {
      setPage(page);
    }
  };

  return children({
    currentItems,
    goToFirstPage,
    goToLastPage,
    goToNextPage,
    goToPreviousPage,
    isFirstPage,
    isLastPage,
    page,
    goToPage,
    pages: [...Array(lastPage).keys()].map((x) => x + 1),
    pagesCount: lastPage,
  });
};

export default PaginationProvider;

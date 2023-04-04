import { useState } from "react";

const PaginationProvider = <TArrayItem,>({
  children,
  items,
  pageSize: pageSizeProp,
}: {
  children: (props: {
    currentItems: TArrayItem[];
    gotoPage: (page: number) => void;
    gotoFirstPage: () => void;
    gotoLastPage: () => void;
    gotoNextPage: () => void;
    gotoPreviousPage: () => void;
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

  const gotoFirstPage = () => setPage(1);
  const gotoLastPage = () => setPage(lastPage);
  const gotoNextPage = () =>
    setPage((page) => (page < lastPage ? page + 1 : page));
  const gotoPreviousPage = () =>
    setPage((page) => (page > firstPage ? page - 1 : page));

  const gotoPage = (page: number) => setPage(page);

  return children({
    currentItems,
    gotoFirstPage,
    gotoLastPage,
    gotoNextPage,
    gotoPreviousPage,
    isFirstPage,
    isLastPage,
    page,
    gotoPage,
    pages: [...Array(lastPage).keys()].map((x) => x + 1),
    pagesCount: lastPage,
  });
};

export default PaginationProvider;

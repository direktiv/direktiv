import { getTotalPages } from "../Pagination/utils";
import { useState } from "react";

const PaginationProvider = <TArrayItem,>({
  children,
  items,
  pageSize = 10,
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
    currentPage: number;
    pagesList: number[];
    totalPages: number;
  }) => JSX.Element;
  pageSize?: number;
  items: TArrayItem[];
}) => {
  const firstPage = 1;
  const [currentPage, setCurrentPage] = useState(firstPage);
  const totalPages = getTotalPages(items.length, pageSize);
  const isLastPage = currentPage === totalPages;
  const isFirstPage = currentPage === firstPage;

  const sliceStart = (currentPage - 1) * pageSize;
  const sliceEnd = sliceStart + pageSize;
  const currentItems = items.slice(sliceStart, sliceEnd);

  // add test for goToPage
  // rename
  const goToFirstPage = () => setCurrentPage(1);
  const goToLastPage = () => setCurrentPage(totalPages);
  const goToNextPage = () =>
    setCurrentPage((page) => Math.min(page + 1, totalPages));
  const goToPreviousPage = () => {
    setCurrentPage((page) => Math.max(page - 1, firstPage));
  };
  const goToPage = (page: number) => {
    if (page >= firstPage && page <= totalPages) {
      setCurrentPage(page);
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
    currentPage,
    goToPage,
    pagesList: [...Array(totalPages).keys()].map((x) => x + 1),
    totalPages,
  });
};

export default PaginationProvider;

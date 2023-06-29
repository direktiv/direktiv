import { Dispatch, SetStateAction } from "react";
import {
  PaginationLink,
  Pagination as PaginationWrapper,
} from "~/design/Pagination";

type SetState<T> = Dispatch<SetStateAction<T>>;

/**
 * *1* 2 ... 9 10
 * 1 *2* 3 ... 9 10
 * 1 2 *3* 4 ... 9 10
 * 1 2 3 *4* 5 ... 9 10
 * 1 2 3 4 *5* 6 ... 9 10
 * 1 2 ... 5 *6* 7 ... 9 10
 */

type PaginationShape = (number | "...")[];

export const generatePaginationPages = ({
  pages,
  currentPage: current,
  neighbours: pageNeighbours = 1,
}: {
  pages: number;
  currentPage: number;
  neighbours?: number;
}): PaginationShape => {
  if (current < 1) return [];
  if (pages < 1) return [];
  if (current > pages) return [];

  /**
   * considering this pagination example
   * 1 2 ... 5 *6* 7 ... 9 10
   *
   * activeStart = 5
   * activeEnd = 7
   * middleSegment = [5, 6, 7]
   * middleSegmentCount = 3
   *
   * startSegment = [1, 2, "..."]
   * endSegment = ["...", 9, 10]
   *
   * and this this function will return
   * [1, 2, "...", 5, 6, 7, "...", 9, 10]
   */

  // middle segment
  const leftCurrent = current - pageNeighbours;
  const activeStart = leftCurrent > 0 ? leftCurrent : current;

  const rightCurrent = current + pageNeighbours;
  const activeEnd = rightCurrent <= pages ? rightCurrent : current;

  const middleSegmentCount = activeEnd - activeStart + 1;

  const middleSegment: PaginationShape = [];
  for (let index = 0; index < middleSegmentCount; index++) {
    middleSegment.push(activeStart + index);
  }

  // left segment
  const leftSegment: PaginationShape = [];

  /**
   * the active segment might also act as the start segment
   * in this case we don't need to generate the left segment
   *  f.e. 1 *2* 3 ... 9 10
   */
  if (activeStart > 1) {
    const startStart = 1;
    let startEnd = startStart + pageNeighbours;
    // remove possible overlap
    if (startEnd >= activeStart) {
      startEnd = activeStart - 1;
    }

    const leftSegmentCount = startEnd - startStart + 1;
    for (let index = 0; index < leftSegmentCount; index++) {
      leftSegment.push(startStart + index);
    }

    // dots needed?
    if (activeStart - startEnd > 1) {
      leftSegment.push("...");
    }
  }

  // right segment
  const rightSegment: PaginationShape = [];

  /**
   * the active segment might also act as the right segment
   * in this case we don't need to generate the rigt segment
   *  f.e. 1 2 ... *9* 10
   */
  if (activeEnd < pages) {
    const endEnd = pages;
    let endStart = endEnd - pageNeighbours;
    // remove possible overlap
    if (endStart <= activeEnd) {
      endStart = activeEnd + 1;
    }

    // dots needed?
    if (endStart - activeEnd > 1) {
      rightSegment.push("...");
    }

    const rightSegmentCount = endEnd - endStart + 1;
    for (let index = 0; index < rightSegmentCount; index++) {
      rightSegment.push(endStart + index);
    }
  }

  return [...leftSegment, ...middleSegment, ...rightSegment];
};

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

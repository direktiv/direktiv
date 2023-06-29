type PaginationShape = (number | "...")[];

/**
 * *1* 2 ... 9 10
 * 1 *2* 3 ... 9 10
 * 1 2 *3* 4 ... 9 10
 * 1 2 3 *4* 5 ... 9 10
 * 1 2 3 4 *5* 6 ... 9 10
 * 1 2 ... 5 *6* 7 ... 9 10
 */
const describePagination = ({
  pages,
  currentPage: current,
  neighbours = 1,
}: {
  pages: number;
  currentPage: number;
  neighbours?: number;
}): PaginationShape => {
  if (current < 1) return [];
  if (pages < 1) return [];
  if (current > pages) return [];
  if (neighbours < 0) return [];

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
  const leftCurrent = current - neighbours;
  const activeStart = leftCurrent > 0 ? leftCurrent : current;

  const rightCurrent = current + neighbours;
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
    let startEnd = startStart + neighbours;
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
    let endStart = endEnd - neighbours;
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

export default describePagination;

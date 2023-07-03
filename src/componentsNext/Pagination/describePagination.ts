type PaginationShape = (number | "...")[];

/**
 *
 * describePagination is a helper function to describe a pagination
 * depending on the current page and the amount of pages. It will
 * return an array of numbers and "...", where the numbers are the
 * pages and the "..." are the dots in between.
 *
 * To handle a high amount of pages, it segments the pagination into
 * three parts: the left segment, the middle segment and the right
 * segment.
 *
 * here is an example of a 10 page pagination with 1 neighbour. the
 * number wrapped in * is the current page
 *
 * *1* 2 ... 9 10
 * 1 *2* 3 ... 9 10
 * 1 2 *3* 4 ... 9 10
 * 1 2 3 *4* 5 ... 9 10
 * 1 2 3 4 *5* 6 ... 9 10
 * 1 2 ... 5 *6* 7 ... 9 10
 *
 * @param pages the amount of pages
 * @param currentPage the page we are currently on
 * @param neighbours the amount of neighbours to the left and right of the current page, start and end defaults to 1
 * @returns an array of numbers and "..."
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
   * the variables will be set as follows
   *
   * pages = 10
   * current = 6
   * neighbours = 1
   *
   * activeSegmentLeft = 5
   * activeSegmentRight = 7
   * activeSegment = [5, 6, 7]
   * activeSegmentCount = 3
   *
   * startSegmentCount = 2
   * startSegmentLeft = 1
   * startSegmentRight = 2
   * startSegment = [1, 2, "..."]
   *
   * endSegmentLeft = 9
   * endSegmentRight = 10
   * endSegmentCount = 2
   * endSegment = ["...", 9, 10]
   *
   * and this this function will return
   * [1, 2, "...", 5, 6, 7, "...", 9, 10]
   */

  // active segment
  const currentLeft = current - neighbours;
  const activeSegmentLeft = currentLeft > 0 ? currentLeft : current;

  const currentRight = current + neighbours;
  const activeSegmentRight = currentRight <= pages ? currentRight : current;

  const activeSegmentCount = activeSegmentRight - activeSegmentLeft + 1;

  const activeSegment: PaginationShape = [];
  for (let index = 0; index < activeSegmentCount; index++) {
    activeSegment.push(activeSegmentLeft + index);
  }

  // start segment
  const startSegment: PaginationShape = [];

  /**
   * the active segment might also act as the start segment
   * in this case we don't need to generate the start segment
   *  f.e. 1 *2* 3 ... 9 10
   */
  if (activeSegmentLeft > 1) {
    const startSegmentLeft = 1;
    let startSegmentRight = startSegmentLeft + neighbours;
    // remove possible overlap
    if (startSegmentRight >= activeSegmentLeft) {
      startSegmentRight = activeSegmentLeft - 1;
    }

    const startSegmentCount = startSegmentRight - startSegmentLeft + 1;
    for (let index = 0; index < startSegmentCount; index++) {
      startSegment.push(startSegmentLeft + index);
    }

    /**
     * analyze the gap between the start segment and the active segment.
     * when there is just one page in between, we don't need to generate
     * the dots and just add the missing page to the start segment
     *
     * when there is more than one page in between, we need to generate
     * the dots
     */
    if (activeSegmentLeft - startSegmentRight === 2) {
      startSegment.push(startSegmentRight + 1);
    }

    if (activeSegmentLeft - startSegmentRight > 2) {
      startSegment.push("...");
    }
  }

  // end segment
  const endSegment: PaginationShape = [];

  /**
   * the active segment might also act as the end segment
   * in this case we don't need to generate the end segment
   *  f.e. 1 2 ... *9* 10
   */
  if (activeSegmentRight < pages) {
    const endSegmentRight = pages;
    let endSegmentLeft = endSegmentRight - neighbours;
    // remove possible overlap
    if (endSegmentLeft <= activeSegmentRight) {
      endSegmentLeft = activeSegmentRight + 1;
    }

    /**
     * analyze the gap between the active segment and the end segment.
     * when there is just one page in between, we don't need to generate
     * the dots and just add the missing page to the end segment
     *
     * when there is more than one page in between, we need to generate
     * the dots
     */
    if (endSegmentLeft - activeSegmentRight === 2) {
      endSegment.push(endSegmentLeft - 1);
    }

    if (endSegmentLeft - activeSegmentRight > 2) {
      endSegment.push("...");
    }

    const endSegmentCount = endSegmentRight - endSegmentLeft + 1;
    for (let index = 0; index < endSegmentCount; index++) {
      endSegment.push(endSegmentLeft + index);
    }
  }

  return [...startSegment, ...activeSegment, ...endSegment];
};

export default describePagination;

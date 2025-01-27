export const getOffsetByPageNumber = (page: number, pageSize: number) =>
  (page - 1) * pageSize;

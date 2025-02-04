export const getOffsetByPageNumber = (page: number, pageSize: number) =>
  (page - 1) * pageSize;

export const getTotalPages = (items: number, pageSize: number) =>
  Math.max(1, Math.ceil(items / pageSize));

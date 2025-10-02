export const parseStringToDate = (value: string): Date | undefined => {
  if (!value) return undefined;
  const parsedDate = new Date(value);
  const isValidDate = isNaN(parsedDate.getTime());
  return isValidDate ? undefined : parsedDate;
};

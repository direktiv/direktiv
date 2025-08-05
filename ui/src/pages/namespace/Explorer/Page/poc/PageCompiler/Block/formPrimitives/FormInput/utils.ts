export const parseStringToDate = (value: string): Date | undefined => {
  if (!value) return undefined;
  const parsedDate = new Date(value);
  return isNaN(parsedDate.getTime()) ? undefined : parsedDate;
};

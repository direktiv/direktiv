export type BlockPath = string;

const blockPathSeparator = ".";

type BlockPathSegment = string | number;

export const addSegmentsToPath = (
  path: BlockPath,
  pathSegment: BlockPathSegment | BlockPathSegment[]
): BlockPath => {
  const pathSegmentArr = Array.isArray(pathSegment)
    ? pathSegment
    : [pathSegment];

  return `${path}${blockPathSeparator}${pathSegmentArr.join(blockPathSeparator)}`;
};

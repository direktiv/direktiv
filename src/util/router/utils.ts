export const analyzePath = (path?: string) => {
  // convert undefined and empty path to null
  let pathClean: string | null = path || null;

  if (path === "/") {
    pathClean = null;
  }

  const segments = pathClean?.split("/") ?? [];

  const segmentsArr = segments.map((s, index, src) => ({
    relative: s,
    absolute: src.slice(0, index + 1).join("/"),
  }));

  return {
    path: pathClean,
    isRoot: segments.length === 0,
    parent: segmentsArr.length > 1 ? segmentsArr[segmentsArr.length - 2] : null,
    segments: segmentsArr,
  };
};

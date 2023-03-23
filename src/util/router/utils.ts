export const analyzePath = (path?: string) => {
  // convert undefined and empty path to null
  let pathClean: string | null = path || null;

  if (path === "/") {
    pathClean = null;
  }

  const segments = pathClean?.split("/") ?? [];
  return {
    path: pathClean,
    segments: segments.map((s, index, src) => ({
      relative: s,
      absolute: src.slice(0, index + 1).join("/"),
    })),
  };
};

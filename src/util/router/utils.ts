import { useMatches } from "react-router-dom";
import { z } from "zod";

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

// react routers useMatches returns an array of match data, this method checks
// for a specific matcher. Please check the corespoinding test for an example
// https://reactrouter.com/en/main/hooks/use-match
export const checkHandlerInMatcher = (
  urlMatcher: ReturnType<typeof useMatches>[0] | undefined,
  handle: string
) =>
  z
    .object({
      handle: z.object({
        [handle]: z.literal(true),
      }),
    })
    .safeParse(urlMatcher).success;

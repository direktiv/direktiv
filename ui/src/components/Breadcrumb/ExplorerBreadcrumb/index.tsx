import { useMatch, useRouterState } from "@tanstack/react-router";

import BreadcrumbSegment from "./BreadcrumbSegment";
import { analyzePath } from "~/util/router/utils";

const ExplorerBreadcrumb = () => {
  const isExplorerPage = useMatch({
    from: "/n/$namespace/explorer",
    shouldThrow: false,
  });
  const { location } = useRouterState();
  const path = analyzePath(location.pathname);

  if (!isExplorerPage) return null;

  return (
    <>
      {path.segments.map((x, i) => (
        <BreadcrumbSegment
          key={x.absolute}
          absolute={x.absolute}
          relative={x.relative}
          isLast={i === path.segments.length - 1}
        />
      ))}
    </>
  );
};

export default ExplorerBreadcrumb;

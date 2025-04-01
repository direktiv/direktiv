import { useMatch, useParams } from "@tanstack/react-router";

import BreadcrumbSegment from "./BreadcrumbSegment";
import { analyzePath } from "~/util/router/utils";

const ExplorerBreadcrumb = () => {
  const isExplorerSubPage = useMatch({
    from: "/n/$namespace/explorer",
    shouldThrow: false,
  });
  const { _splat } = useParams({ strict: false });

  const path = analyzePath(_splat);

  if (!isExplorerSubPage) return null;

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

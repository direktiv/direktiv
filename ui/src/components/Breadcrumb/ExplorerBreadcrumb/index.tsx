import BreadcrumbSegment from "./BreadcrumbSegment";
import { analyzePath } from "~/util/router/utils";
import { usePages } from "~/util/router/pages";

const ExplorerBreadcrumb = () => {
  const pages = usePages();
  const { isExplorerPage, path: pathParams } = pages.explorer.useParams();
  const path = analyzePath(pathParams);

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

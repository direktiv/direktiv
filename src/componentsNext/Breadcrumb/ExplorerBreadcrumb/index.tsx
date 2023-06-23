import BreadcrumbSegment from "./BreadcrumbSegment";
import { analyzePath } from "~/util/router/utils";
import { pages } from "~/util/router/pages";

const ExplorerBreadcrumb = () => {
  const { path: pathParams } = pages.explorer.useParams();
  const path = analyzePath(pathParams);

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

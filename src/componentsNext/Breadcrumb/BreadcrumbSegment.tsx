import { FolderOpen, Play } from "lucide-react";

import { Breadcrumb as BreadcrumbLink } from "../../design/Breadcrumbs";
import { FC } from "react";
import { Link } from "react-router-dom";
import { pages } from "../../util/router/pages";
import { useNamespace } from "../../util/store/namespace";

const BreadcrumbSegment: FC<{
  absolute: string;
  relative: string;
  isLast: boolean;
}> = ({ absolute, relative, isLast }) => {
  const namespace = useNamespace();
  if (!namespace) return null;

  const isWorkflow =
    isLast && (relative.endsWith(".yml") || relative.endsWith(".yaml"));

  const Icon = isWorkflow ? Play : FolderOpen;

  const link = pages.explorer.createHref({
    namespace,
    path: absolute,
    subpage: isWorkflow ? "workflow" : undefined,
  });

  return (
    <BreadcrumbLink>
      <Link to={link} data-testid="breadcrumb-segment">
        <Icon aria-hidden="true" />
        {relative}
      </Link>
    </BreadcrumbLink>
  );
};

export default BreadcrumbSegment;

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

  const { path: pathParamsWorkflow } = pages.workflow.useParams();
  const isWorkflow = !!pathParamsWorkflow && isLast;

  const Icon = isWorkflow ? Play : FolderOpen;

  const link = isWorkflow
    ? pages.workflow.createHref({ namespace, path: absolute })
    : pages.explorer.createHref({ namespace, path: absolute });

  return (
    <BreadcrumbLink>
      <Link to={link}>
        <Icon aria-hidden="true" />
        {relative}
      </Link>
    </BreadcrumbLink>
  );
};

export default BreadcrumbSegment;

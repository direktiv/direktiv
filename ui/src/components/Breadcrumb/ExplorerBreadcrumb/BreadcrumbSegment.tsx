import { fileTypeToExplorerSubpage, fileTypeToIcon } from "~/api/files/utils";

import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { FC } from "react";
import { Link } from "react-router-dom";
import { pages } from "~/util/router/pages";
import { useFile } from "~/api/files/query/file";
import { useNamespace } from "~/util/store/namespace";

const BreadcrumbSegment: FC<{
  absolute: string;
  relative: string;
  isLast: boolean;
}> = ({ absolute, relative, isLast }) => {
  const namespace = useNamespace();
  /**
   * the last breadcrumb item in the file browser can be a file
   * we need to request file information to figure out which
   * icon to use
   */

  const { data } = useFile({
    path: absolute,
    enabled: isLast,
  });

  if (!namespace) return null;
  if (isLast && !data) return null;

  const Icon = fileTypeToIcon(data?.type ?? "directory");

  const link = pages.explorer.createHref({
    namespace,
    path: absolute,
    subpage: fileTypeToExplorerSubpage(data?.type ?? "directory"),
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

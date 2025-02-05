import { fileTypeToExplorerRoute, fileTypeToIcon } from "~/api/files/utils";

import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { FC } from "react";
import { Link } from "@tanstack/react-router";
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

  const route = fileTypeToExplorerRoute(data?.type ?? "directory");

  return (
    <BreadcrumbLink data-testid="breadcrumb-segment">
      <Link to={route} params={{ namespace, _splat: absolute }}>
        <Icon aria-hidden="true" />
        {relative}
      </Link>
    </BreadcrumbLink>
  );
};

export default BreadcrumbSegment;

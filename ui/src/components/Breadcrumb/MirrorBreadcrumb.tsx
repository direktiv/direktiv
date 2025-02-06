import { Link, useMatch, useParams } from "@tanstack/react-router";

import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import CopyButton from "~/design/CopyButton";
import { GitCompare } from "lucide-react";
import { useTranslation } from "react-i18next";

const MirrorBreadcrumb = () => {
  const { sync } = useParams({ strict: false });
  const isMirrorPage = useMatch({
    from: "/n/$namespace/mirror/",
    shouldThrow: false,
  });
  const isSyncDetailPage = useMatch({
    from: "/n/$namespace/mirror/logs/$sync",
    shouldThrow: false,
  });
  const { t } = useTranslation();

  if (!(isMirrorPage || isSyncDetailPage)) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link to="/n/$namespace/mirror" from="/n/$namespace">
          <GitCompare aria-hidden="true" />
          {t("components.breadcrumb.mirror")}
        </Link>
      </BreadcrumbLink>
      {isSyncDetailPage && sync ? (
        <BreadcrumbLink>
          <GitCompare aria-hidden="true" />
          <Link
            to="/n/$namespace/mirror/logs/$sync"
            from="/n/$namespace"
            params={{ sync }}
          >
            {sync.slice(0, 8)}
          </Link>
          <CopyButton
            value={sync}
            buttonProps={{
              variant: "outline",
              className: "hidden group-hover:inline-flex",
              size: "sm",
            }}
          />
        </BreadcrumbLink>
      ) : null}
    </>
  );
};

export default MirrorBreadcrumb;

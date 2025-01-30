import { Link, useMatch, useParams } from "@tanstack/react-router";

import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import CopyButton from "~/design/CopyButton";
import { GitCompare } from "lucide-react";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const MirrorBreadcrumb = () => {
  const namespace = useNamespace();
  const { id } = useParams({ strict: false });
  const isMirrorPage = useMatch({
    from: "/n/$namespace/mirror/",
    shouldThrow: false,
  });
  const isSyncDetailPage = useMatch({
    from: "/n/$namespace/mirror/logs/$id",
    shouldThrow: false,
  });
  const { t } = useTranslation();

  if (!(isMirrorPage || isSyncDetailPage)) return null;
  if (!namespace) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link to="/n/$namespace/mirror" params={{ namespace }}>
          <GitCompare aria-hidden="true" />
          {t("components.breadcrumb.mirror")}
        </Link>
      </BreadcrumbLink>
      {isSyncDetailPage && id ? (
        <BreadcrumbLink>
          <GitCompare aria-hidden="true" />
          <Link to="/n/$namespace/mirror/logs/$id" params={{ namespace, id }}>
            {id.slice(0, 8)}
          </Link>
          <CopyButton
            value={id}
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

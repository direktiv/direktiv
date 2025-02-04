import { Link, useMatch } from "@tanstack/react-router";

import { BadgeCheck } from "lucide-react";
import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import GroupsBreadcrumb from "./GroupsBreadcrumb";
import PolicyBreadcrumb from "./PolicyBreadcumb";
import TokensBreadcrumb from "./TokenBreadCrumb";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const PermissionsBreadcrumb = () => {
  const namespace = useNamespace();
  const { t } = useTranslation();

  const isPermissionsPage = useMatch({
    from: "/n/$namespace/permissions/",
    shouldThrow: false,
  });

  if (!isPermissionsPage) return null;
  if (!namespace) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link to="/n/$namespace/permissions" params={{ namespace }}>
          <BadgeCheck aria-hidden="true" />
          {t("components.breadcrumb.permissions")}
        </Link>
      </BreadcrumbLink>
      <PolicyBreadcrumb />
      <GroupsBreadcrumb />
      <TokensBreadcrumb />
    </>
  );
};

export default PermissionsBreadcrumb;

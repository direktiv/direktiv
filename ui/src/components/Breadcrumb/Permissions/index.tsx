import { Link, useMatch } from "@tanstack/react-router";

import { BadgeCheck } from "lucide-react";
import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import RolesBreadcrumb from "./RolesBreadcrumb";
import TokensBreadcrumb from "./TokenBreadCrumb";
import { useTranslation } from "react-i18next";

const PermissionsBreadcrumb = () => {
  const { t } = useTranslation();

  const isPermissionsPage = useMatch({
    from: "/n/$namespace/permissions",
    shouldThrow: false,
  });

  if (!isPermissionsPage) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link to="/n/$namespace/permissions" from="/n/$namespace">
          <BadgeCheck aria-hidden="true" />
          {t("components.breadcrumb.permissions")}
        </Link>
      </BreadcrumbLink>
      <RolesBreadcrumb />
      <TokensBreadcrumb />
    </>
  );
};

export default PermissionsBreadcrumb;

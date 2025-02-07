import { Link, useMatch } from "@tanstack/react-router";

import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { FileCheck } from "lucide-react";
import { useTranslation } from "react-i18next";

const PolicyBreadcrumb = () => {
  const isPermissionsPolicyPage = useMatch({
    from: "/n/$namespace/permissions/",
    shouldThrow: false,
  });

  const { t } = useTranslation();

  if (!isPermissionsPolicyPage) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link to="/n/$namespace/permissions" from="/n/$namespace">
          <FileCheck aria-hidden="true" />
          {t("components.breadcrumb.permissionsPolicy")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default PolicyBreadcrumb;

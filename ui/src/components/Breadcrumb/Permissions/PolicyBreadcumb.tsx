import { Link, useMatch } from "@tanstack/react-router";

import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { FileCheck } from "lucide-react";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const PolicyBreadcrumb = () => {
  const namespace = useNamespace();
  const isPermissionsPolicyPage = useMatch({
    from: "/n/$namespace/permissions/",
    shouldThrow: false,
  });

  const { t } = useTranslation();

  if (!isPermissionsPolicyPage) return null;
  if (!namespace) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link to="/n/$namespace/permissions" params={{ namespace }}>
          <FileCheck aria-hidden="true" />
          {t("components.breadcrumb.permissionsPolicy")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default PolicyBreadcrumb;

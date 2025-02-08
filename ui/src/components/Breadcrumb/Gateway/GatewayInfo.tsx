import { Link, useParams } from "@tanstack/react-router";

import { BookOpen } from "lucide-react";
import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { useTranslation } from "react-i18next";

const GatewayInfoBreadcrumb = () => {
  const { _splat: isGatewayInfoPage } = useParams({ strict: false });

  const { t } = useTranslation();

  if (!isGatewayInfoPage) return null;

  return (
    <>
      <BreadcrumbLink data-testid="breadcrumb-info">
        <Link to="/n/$namespace/gateway" from="/n/$namespace">
          <BookOpen aria-hidden="true" />
          {t("components.breadcrumb.gatewayInfo")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default GatewayInfoBreadcrumb;

import { Link, useMatch } from "@tanstack/react-router";

import { ActivitySquare } from "lucide-react";
import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { useTranslation } from "react-i18next";

const MonitoringBreadcrumb = () => {
  const isMonitoringPage = useMatch({
    from: "/n/$namespace/monitoring",
    shouldThrow: false,
  });

  const { t } = useTranslation();

  if (!isMonitoringPage) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link to="/n/$namespace/monitoring" from="/n/$namespace">
          <ActivitySquare aria-hidden="true" />
          {t("components.breadcrumb.monitoring")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default MonitoringBreadcrumb;

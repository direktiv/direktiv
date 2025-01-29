import { Link, useMatch } from "@tanstack/react-router";

import { ActivitySquare } from "lucide-react";
import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const MonitoringBreadcrumb = () => {
  const namespace = useNamespace();

  const isMonitoringPage = useMatch({
    from: "/n/$namespace/monitoring",
    shouldThrow: false,
  });

  const { t } = useTranslation();

  if (!isMonitoringPage) return null;
  if (!namespace) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link to="/n/$namespace/monitoring" params={{ namespace }}>
          <ActivitySquare aria-hidden="true" />
          {t("components.breadcrumb.monitoring")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default MonitoringBreadcrumb;

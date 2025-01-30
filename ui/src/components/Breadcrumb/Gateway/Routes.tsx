import { Link, useMatch, useParams } from "@tanstack/react-router";
import { SquareGanttChartIcon, Workflow } from "lucide-react";

import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const RoutesBreadcrumb = () => {
  const namespace = useNamespace();
  const { filename } = useParams({ strict: false });
  const isGatewayRoutesPage = useMatch({
    from: "/n/$namespace/gateway/routes/",
    shouldThrow: false,
  });

  const isGatewayRoutesDetailPage = useMatch({
    from: "/n/$namespace/gateway/routes/$filename",
    shouldThrow: false,
  });

  const { t } = useTranslation();

  if (!namespace) return null;
  if (!isGatewayRoutesPage && !isGatewayRoutesDetailPage) return null;

  return (
    <>
      <BreadcrumbLink data-testid="breadcrumb-routes">
        <Link to="/n/$namespace/gateway/routes" params={{ namespace }}>
          <Workflow aria-hidden="true" />
          {t("components.breadcrumb.gatewayRoutes")}
        </Link>
      </BreadcrumbLink>
      {filename && (
        <BreadcrumbLink data-testid="breadcrumb-routes">
          <Link
            to="/n/$namespace/gateway/routes/$filename"
            params={{ namespace, filename }}
          >
            <SquareGanttChartIcon aria-hidden="true" />
            {filename}
          </Link>
        </BreadcrumbLink>
      )}
    </>
  );
};

export default RoutesBreadcrumb;

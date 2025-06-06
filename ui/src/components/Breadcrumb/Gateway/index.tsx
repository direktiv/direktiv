import { Link, useMatches } from "@tanstack/react-router";

import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import ConsumerBreadcrumb from "./Consumer";
import { FileRoutesById } from "~/routeTree.gen";
import GatewayDocumentationBreadcrumb from "./Documentation";
import GatewayInfoBreadcrumb from "./GatewayInfo";
import { Network } from "lucide-react";
import RoutesBreadcrumb from "./Routes";
import { useTranslation } from "react-i18next";

const GatewayBreadcrumb = () => {
  const matches = useMatches();
  const routeId: keyof FileRoutesById = "/n/$namespace/gateway";
  const isGatewayPage = matches.some((match) =>
    match.routeId.startsWith(routeId)
  );

  const { t } = useTranslation();

  if (!isGatewayPage) return null;

  return (
    <>
      <BreadcrumbLink data-testid="breadcrumb-gateway">
        <Link to="/n/$namespace/gateway/gatewayInfo" from="/n/$namespace">
          <Network aria-hidden="true" />
          {t("components.breadcrumb.gateway")}
        </Link>
      </BreadcrumbLink>
      <RoutesBreadcrumb />
      <ConsumerBreadcrumb />
      <GatewayInfoBreadcrumb />
      <GatewayDocumentationBreadcrumb />
    </>
  );
};

export default GatewayBreadcrumb;

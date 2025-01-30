import { Link, useMatches } from "@tanstack/react-router";

import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import ConsumerBreadcrumb from "./Consumer";
import { FileRoutesById } from "~/routeTree.gen";
import { Network } from "lucide-react";
import RoutesBreadcrumb from "./Routes";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const GatewayBreadcrumb = () => {
  const namespace = useNamespace();
  const matches = useMatches();
  const routeId: keyof FileRoutesById = "/n/$namespace/gateway/";
  const isGatewayPage = matches.some((match) =>
    match.routeId.startsWith(routeId)
  );

  const { t } = useTranslation();

  if (!namespace) return null;
  if (!isGatewayPage) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link to="/n/$namespace/gateway" params={{ namespace }}>
          <Network aria-hidden="true" />
          {t("components.breadcrumb.gateway")}
        </Link>
      </BreadcrumbLink>
      <RoutesBreadcrumb />
      <ConsumerBreadcrumb />
    </>
  );
};

export default GatewayBreadcrumb;

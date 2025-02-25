import { Link, useMatch } from "@tanstack/react-router";

import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { ScrollText } from "lucide-react";
import { useTranslation } from "react-i18next";

const GatewayDocumentationBreadcrumb = () => {
  const isGatewayDocumentationPage = useMatch({
    from: "/n/$namespace/gateway/openapiDoc",
    shouldThrow: false,
  });
  const { t } = useTranslation();

  if (!isGatewayDocumentationPage) return null;

  return (
    <>
      <BreadcrumbLink data-testid="breadcrumb-documentation">
        <Link to="/n/$namespace/gateway/openapiDoc" from="/n/$namespace">
          <ScrollText aria-hidden="true" />
          {t("components.breadcrumb.gatewayDocumentation")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default GatewayDocumentationBreadcrumb;

import { Link, useMatch } from "@tanstack/react-router";

import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { Users } from "lucide-react";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const ConsumerBreadcrumb = () => {
  const namespace = useNamespace();
  const isGatewayConsumerPage = useMatch({
    from: "/n/$namespace/gateway/consumers",
    shouldThrow: false,
  });
  const { t } = useTranslation();

  if (!namespace) return null;
  if (!isGatewayConsumerPage) return null;

  return (
    <>
      <BreadcrumbLink data-testid="breadcrumb-consumers">
        <Link to="/n/$namespace/gateway/consumers" params={{ namespace }}>
          <Users aria-hidden="true" />
          {t("components.breadcrumb.gatewayConsumers")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default ConsumerBreadcrumb;

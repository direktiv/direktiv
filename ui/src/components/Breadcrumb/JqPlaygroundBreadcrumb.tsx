import { Link, useMatch } from "@tanstack/react-router";

import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { PlaySquare } from "lucide-react";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const JqPlaygroundBreadcrumb = () => {
  const namespace = useNamespace();
  const isJqPlaygroundPage = useMatch({
    from: "/n/$namespace/jq",
    shouldThrow: false,
  });
  const { t } = useTranslation();

  if (!isJqPlaygroundPage) return null;
  if (!namespace) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link to="/n/$namespace/jq" params={{ namespace }}>
          <PlaySquare aria-hidden="true" />
          {t("components.breadcrumb.jqPlayground")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default JqPlaygroundBreadcrumb;

import { Link, useMatch } from "@tanstack/react-router";

import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { PlaySquare } from "lucide-react";
import { useTranslation } from "react-i18next";

const JqPlaygroundBreadcrumb = () => {
  const isJqPlaygroundPage = useMatch({
    from: "/n/$namespace/jq",
    shouldThrow: false,
  });
  const { t } = useTranslation();

  if (!isJqPlaygroundPage) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link to="/n/$namespace/jq" from="/n/$namespace">
          <PlaySquare aria-hidden="true" />
          {t("components.breadcrumb.jqPlayground")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default JqPlaygroundBreadcrumb;

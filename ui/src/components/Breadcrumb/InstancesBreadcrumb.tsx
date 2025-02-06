import { Box, Boxes } from "lucide-react";
import { Link, useMatch, useParams } from "@tanstack/react-router";

import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import CopyButton from "~/design/CopyButton";
import { useTranslation } from "react-i18next";

const InstancesBreadcrumb = () => {
  const { id } = useParams({ strict: false });
  const isInstancePage = useMatch({
    from: "/n/$namespace/instances/",
    shouldThrow: false,
  });
  const isInstanceDetailPage = useMatch({
    from: "/n/$namespace/instances/$id",
    shouldThrow: false,
  });
  const { t } = useTranslation();

  if (!(isInstancePage || isInstanceDetailPage)) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link to="/n/$namespace/instances" from="/n/$namespace">
          <Boxes aria-hidden="true" />
          {t("components.breadcrumb.instances")}
        </Link>
      </BreadcrumbLink>
      {isInstanceDetailPage && id ? (
        <BreadcrumbLink>
          <Box aria-hidden="true" />
          <Link
            to="/n/$namespace/instances/$id"
            from="/n/$namespace"
            params={{ id }}
          >
            {id.slice(0, 8)}
          </Link>
          <CopyButton
            value={id}
            buttonProps={{
              variant: "outline",
              className: "hidden group-hover:inline-flex",
              size: "sm",
            }}
          />
        </BreadcrumbLink>
      ) : null}
    </>
  );
};

export default InstancesBreadcrumb;

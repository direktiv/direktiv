import { Box, Boxes } from "lucide-react";
import { Link, useMatch, useParams } from "@tanstack/react-router";

import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import CopyButton from "~/design/CopyButton";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const InstancesBreadcrumb = () => {
  const namespace = useNamespace();
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
  if (!namespace) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link to="/n/$namespace/instances" params={{ namespace }}>
          <Boxes aria-hidden="true" />
          {t("components.breadcrumb.instances")}
        </Link>
      </BreadcrumbLink>
      {isInstanceDetailPage && id ? (
        <BreadcrumbLink>
          <Box aria-hidden="true" />
          <Link to="/n/$namespace/instances/$id" params={{ namespace, id }}>
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

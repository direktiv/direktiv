import { TableCell, TableRow } from "~/design/Table";

import { AllowAnonymous } from "./Anonymous";
import Badge from "~/design/Badge";
import { FC } from "react";
import { Link } from "react-router-dom";
import MessagesOverlay from "./MessagesOverlay";
import { Methods } from "./Methods";
import Plugins from "./Plugins";
import PublicPathInput from "./PublicPath";
import { RouteSchemaType } from "~/api/gateway/schema";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

type RowProps = {
  gateway: RouteSchemaType;
};

export const Row: FC<RowProps> = ({ gateway }) => {
  const namespace = useNamespace();
  const { t } = useTranslation();
  if (!namespace) return null;

  const path = gateway.server_path
    ? `${window.location.origin}${gateway.server_path}`
    : undefined;

  return (
    <TableRow>
      <TableCell>
        <div className="flex flex-col gap-3">
          <Link
            className="whitespace-normal break-all hover:underline"
            to={pages.explorer.createHref({
              namespace,
              path: gateway.file_path,
              subpage: "endpoint",
            })}
          >
            {gateway.file_path}
          </Link>
          <div className="flex gap-1">
            <MessagesOverlay messages={gateway.errors} variant="error">
              {(errorCount) => (
                <Badge variant="destructive">
                  {t("pages.gateway.routes.row.error.count", {
                    count: errorCount,
                  })}
                </Badge>
              )}
            </MessagesOverlay>
            <MessagesOverlay messages={gateway.warnings} variant="warning">
              {(warningCount) => (
                <Badge variant="secondary">
                  {t("pages.gateway.routes.row.warnings.count", {
                    count: warningCount,
                  })}
                </Badge>
              )}
            </MessagesOverlay>
          </div>
        </div>
      </TableCell>
      <TableCell>
        <Methods methods={gateway.methods} />
      </TableCell>
      <TableCell className="whitespace-normal break-all">
        {path && <PublicPathInput path={path} />}
      </TableCell>
      <TableCell>
        <Plugins plugins={gateway.plugins} />
      </TableCell>
      <TableCell>
        <AllowAnonymous allow={gateway.allow_anonymous} />
      </TableCell>
    </TableRow>
  );
};

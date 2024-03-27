import { Link, useNavigate } from "react-router-dom";
import { TableCell, TableRow } from "~/design/Table";

import { AllowAnonymous } from "./Anonymous";
import Badge from "~/design/Badge";
import { FC } from "react";
import MessagesOverlay from "./MessagesOverlay";
import { Methods } from "./Methods";
import Plugins from "./Plugins";
import PublicPathInput from "./PublicPath";
import { RouteSchemaType } from "~/api/gateway/schema";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

type RowProps = {
  route: RouteSchemaType;
};

export const Row: FC<RowProps> = ({ route }) => {
  const namespace = useNamespace();
  const { t } = useTranslation();
  const navigate = useNavigate();
  if (!namespace) return null;

  return (
    <TableRow
      onClick={() => {
        navigate(
          pages.gateway.createHref({
            namespace,
            subpage: "routeDetail",
            routePath: route.file_path,
          })
        );
      }}
      className="cursor-pointer"
    >
      <TableCell>
        <div className="flex flex-col gap-3">
          <Link
            onClick={(e) => {
              e.stopPropagation(); // prevent the onClick on the row from firing when clicking the workflow link
            }}
            className="whitespace-normal break-all hover:underline"
            to={pages.explorer.createHref({
              namespace,
              path: route.file_path,
              subpage: "endpoint",
            })}
          >
            {route.file_path}
          </Link>
          <div className="flex gap-1">
            <MessagesOverlay messages={route.errors} variant="error">
              {(errorCount) => (
                <Badge variant="destructive">
                  {t("pages.gateway.routes.row.error.count", {
                    count: errorCount,
                  })}
                </Badge>
              )}
            </MessagesOverlay>
            <MessagesOverlay messages={route.warnings} variant="warning">
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
        <Methods methods={route.methods} />
      </TableCell>
      <TableCell className="whitespace-normal break-all">
        {route.server_path && <PublicPathInput path={route.server_path} />}
      </TableCell>
      <TableCell>
        <Plugins plugins={route.plugins} />
      </TableCell>
      <TableCell>
        <AllowAnonymous allow={route.allow_anonymous} />
      </TableCell>
    </TableRow>
  );
};

import { Link, useNavigate } from "@tanstack/react-router";
import { TableCell, TableRow } from "~/design/Table";

import Badge from "~/design/Badge";
import { FC } from "react";
import MessagesOverlay from "../components/MessagesOverlay";
import { Methods } from "../components/Methods";
import Plugins from "../components/Plugins";
import PublicPathInput from "../components/PublicPath";
import { RouteSchemaType } from "~/api/gateway/schema";
import { getMethodsFromOpenApiSpec } from "../utils";
import { useTranslation } from "react-i18next";

type RowProps = {
  route: RouteSchemaType;
};

export const Row: FC<RowProps> = ({ route }) => {
  const { t } = useTranslation();
  const navigate = useNavigate();

  return (
    <TableRow
      onClick={() => {
        navigate({
          to: "/n/$namespace/gateway/routes/$",
          from: "/n/$namespace",
          params: { _splat: route.file_path },
        });
      }}
      className="cursor-pointer"
    >
      <TableCell>
        <div className="flex grow my-1">
          <Link
            onClick={(e) => {
              e.stopPropagation(); // prevent the onClick on the row from firing when clicking the workflow link
            }}
            className="whitespace-normal hover:underline"
            to="/n/$namespace/explorer/endpoint/$"
            from="/n/$namespace"
            params={{ _splat: route.file_path }}
          >
            {route.file_path}
          </Link>
        </div>
        <div className="flex flex-row items-start gap-1">
          {/* badges */}
          <MessagesOverlay messages={route.warnings} variant="warning">
            {(warningCount) => (
              <Badge variant="secondary">
                {t("pages.gateway.routes.row.warnings.count", {
                  count: warningCount,
                })}
              </Badge>
            )}
          </MessagesOverlay>
          <Methods methods={getMethodsFromOpenApiSpec(route.spec)} />
          {route.spec["x-direktiv-config"]?.allow_anonymous && (
            <Badge variant="outline">
              {t("pages.gateway.routes.row.allowAnonymous.public")}
            </Badge>
          )}

          <MessagesOverlay messages={route.errors} variant="error">
            {(errorCount) => (
              <Badge variant="destructive">
                {t("pages.gateway.routes.row.error.count", {
                  count: errorCount,
                })}
              </Badge>
            )}
          </MessagesOverlay>
        </div>
      </TableCell>
      {/* badges end */}

      <TableCell>
        <Plugins plugins={route.spec["x-direktiv-config"]?.plugins} />
      </TableCell>
      <TableCell className="whitespace-normal break-all">
        {route.server_path && <PublicPathInput path={route.server_path} />}
      </TableCell>
    </TableRow>
  );
};

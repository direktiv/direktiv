import { Link, useNavigate } from "@tanstack/react-router";
import { TableCell, TableRow } from "~/design/Table";

import { AllowAnonymous } from "../components/Anonymous";
import Badge from "~/design/Badge";
import { FC } from "react";
import MessagesOverlay from "../components/MessagesOverlay";
import { Methods } from "../components/Methods";
import Plugins from "../components/Plugins";
import PublicPathInput from "../components/PublicPath";
import { RouteSchemaType } from "~/api/gateway/schema";
import { getMethodFromOpenApiSpec } from "../utils";
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
          params: { _splat: route?.file_path },
        });
      }}
      className="cursor-pointer"
    >
      <TableCell>
        <div className="flex flex-col items-start gap-3">
          <Link
            onClick={(e) => {
              e.stopPropagation(); // prevent the onClick on the row from firing when clicking the workflow link
            }}
            className="whitespace-normal break-all hover:underline"
            to="/n/$namespace/explorer/endpoint/$"
            from="/n/$namespace"
            params={{ _splat: route?.file_path }}
          >
            {route?.file_path}
          </Link>
          <div className="flex gap-1">
            <MessagesOverlay messages={route?.errors ?? []} variant="error">
              {(errorCount) => (
                <Badge variant="destructive">
                  {t("pages.gateway.routes.row.error.count", {
                    count: errorCount,
                  })}
                </Badge>
              )}
            </MessagesOverlay>
            <MessagesOverlay messages={route?.warnings ?? []} variant="warning">
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
        <Methods methods={getMethodFromOpenApiSpec(route.spec)} />
      </TableCell>
      <TableCell>
        <Plugins plugins={route?.spec["x-direktiv-config"]?.plugins ?? {}} />
      </TableCell>
      <TableCell>
        <AllowAnonymous
          allow={route?.spec["x-direktiv-config"]?.allow_anonymous ?? false}
        />
      </TableCell>
      <TableCell className="whitespace-normal break-all">
        {route?.server_path && <PublicPathInput path={route?.server_path} />}
      </TableCell>
    </TableRow>
  );
};

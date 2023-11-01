import { TableCell, TableRow } from "~/design/Table";

import Badge from "~/design/Badge";
import ErrorBadge from "./ErrorBadge";
import { FC } from "react";
import { GatewaySchemeType } from "~/api/gateway/schema";
import { useTranslation } from "react-i18next";

type RowProps = {
  gateway: GatewaySchemeType;
};

export const Row: FC<RowProps> = ({ gateway }) => {
  const numberOfPlugins = gateway.plugins.length;
  const { t } = useTranslation();
  return (
    <TableRow>
      <TableCell>
        {gateway.file_path} <ErrorBadge error={gateway.error} />
      </TableCell>
      <TableCell>
        <Badge variant="secondary">{gateway.method}</Badge>
      </TableCell>
      <TableCell>
        {t("pages.gateway.row.plugins", { count: numberOfPlugins })}
      </TableCell>
    </TableRow>
  );
};

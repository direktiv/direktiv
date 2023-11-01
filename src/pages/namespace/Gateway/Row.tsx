import { TableCell, TableRow } from "~/design/Table";

import Badge from "~/design/Badge";
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
      <TableCell>{gateway.file_path}</TableCell>
      <TableCell>
        <Badge variant="secondary">{gateway.method}</Badge>
      </TableCell>
      <TableCell>
        {t("pages.gateway.row.plugins", { count: numberOfPlugins })}
      </TableCell>
    </TableRow>
  );
};

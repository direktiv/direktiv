import { TableCell, TableRow } from "~/design/Table";

import Badge from "~/design/Badge";
import ErrorBadge from "./ErrorBadge";
import { FC } from "react";
import { GatewaySchemeType } from "~/api/gateway/schema";
import PluginPopover from "./PluginPopover";

type RowProps = {
  gateway: GatewaySchemeType;
};

export const Row: FC<RowProps> = ({ gateway }) => (
  <TableRow>
    <TableCell>
      {gateway.file_path} <ErrorBadge error={gateway.error} />
    </TableCell>
    <TableCell>
      {gateway.method ? (
        <Badge variant="secondary">{gateway.method}</Badge>
      ) : null}
    </TableCell>
    <TableCell>
      <PluginPopover gateway={gateway} />
    </TableCell>
  </TableRow>
);

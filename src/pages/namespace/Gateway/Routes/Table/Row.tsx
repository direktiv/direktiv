import { TableCell, TableRow } from "~/design/Table";

import Badge from "~/design/Badge";
import ErrorBadge from "./ErrorBadge";
import { FC } from "react";
import { GatewaySchemeType } from "~/api/gateway/schema";
import { Link } from "react-router-dom";
import PluginPopover from "./PluginPopover";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";

type RowProps = {
  gateway: GatewaySchemeType;
};

export const Row: FC<RowProps> = ({ gateway }) => {
  const namespace = useNamespace();
  if (!namespace) return null;
  return (
    <TableRow>
      <TableCell>
        <Link
          className="hover:underline"
          to={pages.explorer.createHref({
            namespace,
            path: gateway.file_path,
            subpage: "gateway",
          })}
        >
          {gateway.file_path}
        </Link>{" "}
        <ErrorBadge error={gateway.error} />
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
};

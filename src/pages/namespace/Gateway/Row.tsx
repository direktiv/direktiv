import { TableCell, TableRow } from "~/design/Table";

import { FC } from "react";
import { GatewaySchemeType } from "~/api/gateway/schema";

type RowProps = {
  gateway: GatewaySchemeType;
};

const Row: FC<RowProps> = ({ gateway }) => (
  <TableRow>
    <TableCell>{gateway.file_path}</TableCell>
  </TableRow>
);

export default Row;

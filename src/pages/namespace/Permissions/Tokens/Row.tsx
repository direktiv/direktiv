import { TableCell, TableRow } from "~/design/Table";

import PermissionsInfo from "../components/PermissionsInfo";
import { TokenSchemaType } from "~/api/enterprise/tokens/schema";

const Row = ({ token }: { token: TokenSchemaType }) => (
  <TableRow>
    <TableCell>{token.description}</TableCell>
    <TableCell>
      <PermissionsInfo permissions={token.permissions} />
    </TableCell>
    <TableCell>{token.created}</TableCell>
    <TableCell>{token.expires}</TableCell>
  </TableRow>
);

export default Row;

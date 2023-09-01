import { TableCell, TableRow } from "~/design/Table";

import PermissionsInfo from "../components/PermissionsInfo";
import { TokenSchemaType } from "~/api/enterprise/tokens/schema";

const Row = ({ token }: { token: TokenSchemaType }) => (
  <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
    <TableCell>{token.description}</TableCell>
    <TableCell>{token.created}</TableCell>
    <TableCell>{token.expires}</TableCell>
    <TableCell>
      <PermissionsInfo permissions={token.permissions} />
    </TableCell>
  </TableRow>
);

export default Row;

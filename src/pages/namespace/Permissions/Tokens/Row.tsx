import { TableCell, TableRow } from "~/design/Table";

import { TokenSchemaType } from "~/api/enterprise/tokens/schema";

const Row = ({ token }: { token: TokenSchemaType }) => (
  <TableRow>
    <TableCell>{token.description}</TableCell>
    <TableCell>All permissions</TableCell>
    <TableCell>{token.created}</TableCell>
    <TableCell>{token.expires}</TableCell>
  </TableRow>
);

export default Row;

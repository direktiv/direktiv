import { TableCell, TableRow } from "~/design/Table";

import type { MirrorActivitySchemaType } from "~/api/tree/schema";

const Row = ({ item }: { item: MirrorActivitySchemaType }) => (
  <TableRow>
    <TableCell>{item.status}</TableCell>
    <TableCell>{item.type}</TableCell>
    <TableCell>{item.id}</TableCell>
    <TableCell>{item.createdAt}</TableCell>
  </TableRow>
);

export default Row;

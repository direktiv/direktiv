import { TableCell, TableRow } from "~/design/Table";

import { GroupSchemaType } from "~/api/enterprise/groups/schema";

const Row = ({ group }: { group: GroupSchemaType }) => (
  <TableRow>
    <TableCell>{group.group}</TableCell>
    <TableCell>{group.description}</TableCell>
    <TableCell>All permissions</TableCell>
  </TableRow>
);

export default Row;

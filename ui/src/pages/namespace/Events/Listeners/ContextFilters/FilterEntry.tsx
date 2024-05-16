import { TableCell, TableRow } from "~/design/Table";

import { EventContextFilterSchemaType } from "~/api/eventListeners/schema";

const FilterEntry = ({ filter }: { filter: EventContextFilterSchemaType }) => (
  <>
    <TableRow>
      <TableCell>{filter.type}</TableCell>
    </TableRow>
    {Object.entries(filter.context).map(([key, value]) => (
      <TableRow key={key}>
        <TableCell className="pl-10 text-gray-10">
          {key}: {value}
        </TableCell>
      </TableRow>
    ))}
  </>
);

export default FilterEntry;

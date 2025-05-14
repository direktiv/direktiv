import { TableCell as TableCellDesignComponent } from "~/design/Table";
import { TableColumnType } from "../../../schema/blocks/table/tableColumn";
import { TemplateString } from "../../primitives/TemplateString";

type TableCellProps = {
  blockProps: TableColumnType;
};

export const TableCell = ({ blockProps }: TableCellProps) => {
  const { content } = blockProps;
  return (
    <TableCellDesignComponent className="p-5">
      <TemplateString value={content} />
    </TableCellDesignComponent>
  );
};

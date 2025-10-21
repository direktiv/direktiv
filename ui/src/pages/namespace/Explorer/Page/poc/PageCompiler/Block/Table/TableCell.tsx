import { TableCell as TableCellDesignComponent } from "~/design/Table";
import { TableColumnType } from "../../../schema/blocks/table/tableColumn";
import { TemplateString } from "../../primitives/TemplateString";
import { forwardRef } from "react";

type TableCellProps = {
  blockProps: TableColumnType;
} & Omit<React.HTMLProps<HTMLTableCellElement>, "children" | "ref">;

export const TableCell = forwardRef<HTMLTableCellElement, TableCellProps>(
  ({ blockProps, ...props }, ref) => {
    const { content } = blockProps;
    return (
      <TableCellDesignComponent ref={ref} {...props}>
        <TemplateString value={content} />
      </TableCellDesignComponent>
    );
  }
);

TableCell.displayName = "TableCell";

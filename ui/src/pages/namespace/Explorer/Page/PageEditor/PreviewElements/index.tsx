import { FC, PropsWithChildren } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";

import { Card } from "~/design/Card";
import { ImageIcon } from "lucide-react";

export const DragAndDropPreview: FC<PropsWithChildren> = ({ children }) => (
  <div>{children}</div>
);

export const getElementComponent = (
  element: string,
  hidden: boolean,
  content: string
) => {
  switch (element) {
    case "Header":
      return <DefaultHeader hidden={hidden} content={content} />;
    case "Footer":
      return <DefaultFooter hidden={hidden} content={content} />;
    case "Image":
      return <DefaultImage hidden={hidden} content={content} />;
    case "Text":
      return <DefaultText hidden={hidden} content={content} />;
    case "Table":
      return <DefaultTable hidden={hidden} content={content} />;
    case "List":
      return <DefaultList hidden={hidden} content={content} />;
    default:
      return <></>;
  }
};

type TableProps = {
  columns?: number;
  rows?: number;
};

type previewElementProps = {
  content: string;
  hidden: boolean;
};

const DefaultTable: FC<TableProps & previewElementProps> = (content) => (
  <Table className="p-2 my-2 border-2" hidden={content.hidden}>
    <TableHead className="border-2">
      <TableRow className="hover:bg-transparent">
        <TableHeaderCell>TableHeader1</TableHeaderCell>
        <TableHeaderCell>TableHeader2</TableHeaderCell>
        <TableHeaderCell>TableHeader3</TableHeaderCell>
        <TableHeaderCell>TableHeader4</TableHeaderCell>
        <TableHeaderCell>
          <span className="sr-only">Edit</span>
        </TableHeaderCell>
      </TableRow>
    </TableHead>
    <TableBody>
      <TableRow className="border-2 hover:bg-transparent">
        <TableCell>TableCell 1</TableCell>
        <TableCell>TableCell 2</TableCell>
        <TableCell>TableCell 3</TableCell>
        <TableCell>TableCell 4</TableCell>
      </TableRow>
    </TableBody>
  </Table>
);

const DefaultList: FC<previewElementProps> = (content) => (
  <ul hidden={content.hidden} className="p-2 list-disc pl-4">
    <li>Item 1</li>
    <li>Item 2</li>
    <li>Item 3</li>
  </ul>
);

const DefaultImage: FC<previewElementProps> = (content) => (
  <div hidden={content.hidden} className="p-2">
    <ImageIcon />
  </div>
);

const DefaultText: FC<previewElementProps> = (content) => (
  <p hidden={content.hidden} className="py-2">
    {content.content === undefined ? (
      <>Placeholder Text</>
    ) : (
      <>{content.content}</>
    )}
  </p>
);

const DefaultHeader: FC<previewElementProps> = (content) => (
  <Card noShadow className="border-2 my-2" hidden={content.hidden}>
    <p className="text-lg font-semibold p-2">{content.content}</p>
  </Card>
);

const DefaultFooter: FC<previewElementProps> = (content) => (
  <Card noShadow className="border-2 my-2" hidden={content.hidden}>
    <p className="text-lg font-semibold p-2">{content.content}</p>
  </Card>
);

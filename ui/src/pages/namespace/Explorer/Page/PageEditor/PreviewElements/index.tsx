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
  content: string | tableData
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
    default:
      return <></>;
  }
};

type TableProps = {
  columns?: number;
  rows?: number;
};

type tableData = [
  {
    header: string;
    cell: string;
  },
];

type previewElementProps = {
  content: string | tableData;
  hidden: boolean;
};

const DefaultTable: FC<TableProps & previewElementProps> = (content) => {
  const placeholderData = [
    {
      header: "TableHeader 1",
      cell: "TableCell 1",
    },
    { header: "TableHeader 1", cell: "TableCell 2" },
  ];

  const data =
    typeof content.content !== "string" ? content.content : placeholderData;

  // const data = placeholderData;

  return (
    <Table className="p-2 my-2 border-2 text-xs" hidden={content.hidden}>
      <TableHead className="border-2">
        <TableRow className="hover:bg-transparent">
          {data.map((element, index) => (
            <TableHeaderCell key={index}>{element.header}</TableHeaderCell>
          ))}
        </TableRow>
      </TableHead>
      <TableBody>
        <TableRow className="border-2 hover:bg-transparent">
          {data.map((element, index) => (
            <TableCell key={index}>{element.cell}</TableCell>
          ))}
        </TableRow>
      </TableBody>
    </Table>
  );
};

const DefaultImage: FC<previewElementProps> = (content) => (
  <div hidden={content.hidden} className="p-2">
    {content.content === undefined ? (
      <ImageIcon />
    ) : (
      <img src={content.content} />
    )}
  </div>
);

const DefaultText: FC<previewElementProps> = (content) => (
  <p hidden={content.hidden} className="py-2">
    {typeof content.content !== "string" ? (
      <>Placeholder Text</>
    ) : (
      <>{content.content}</>
    )}
  </p>
);

const DefaultHeader: FC<previewElementProps> = (content) => (
  <Card noShadow className="border-2 my-2" hidden={content.hidden}>
    <p className="text-lg font-semibold p-2">
      {JSON.stringify(content.content)}
    </p>
  </Card>
);

const DefaultFooter: FC<previewElementProps> = (content) => (
  <Card noShadow className="border-2 my-2" hidden={content.hidden}>
    <p className="text-lg font-semibold p-2">
      {JSON.stringify(content.content)}
    </p>
  </Card>
);

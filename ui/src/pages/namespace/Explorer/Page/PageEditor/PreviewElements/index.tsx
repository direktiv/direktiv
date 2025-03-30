import { FC, PropsWithChildren } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";

export const DragAndDropPreview: FC<PropsWithChildren> = ({ children }) => (
  <div>{children}</div>
);

export const getElementComponent = (
  element: string,
  hidden: boolean,
  content: contentType
) => {
  switch (element) {
    case "Header":
      return <DefaultHeader hidden={hidden} content={content} />;
    case "Footer":
      return <DefaultFooter hidden={hidden} content={content} />;
    case "Text":
      return <DefaultText hidden={hidden} content={content} />;
    case "Table":
      return <DefaultTable hidden={hidden} content={content.content} />;
    default:
      return <></>;
  }
};

type contentType = {
  content?: string;
  tableData?: tableData;
};

type tableData = {
  header: string;
  cell: string;
}[];

type previewTableProps = {
  content: tableData | string | undefined;
  hidden: boolean;
};

type previewElementProps = {
  content: contentType;
  hidden: boolean;
};

const DefaultTable: FC<previewTableProps> = ({ content, hidden }) => {
  const placeholderData: tableData = [
    {
      header: "Table Header 1",
      cell: "- no data -",
    },
  ];

  const data = content && Array.isArray(content) ? content : placeholderData;

  return (
    <Table className="p-2 my-2 border-2 text-xs" hidden={hidden}>
      <TableHead className="border-2">
        <TableRow className="hover:bg-transparent">
          {data?.map((element, index) => (
            <TableHeaderCell key={index}>{element.header}</TableHeaderCell>
          ))}
        </TableRow>
      </TableHead>
      <TableBody>
        <TableRow className="border-2 hover:bg-transparent">
          {data?.map((element, index) => (
            <TableCell key={index}>{element.cell}</TableCell>
          ))}
        </TableRow>
      </TableBody>
    </Table>
  );
};

const DefaultText: FC<previewElementProps> = ({ content, hidden }) => (
  <p hidden={hidden} className="py-2">
    {typeof content !== "string" ? <>Placeholder Text</> : <>{content}</>}
  </p>
);

const DefaultHeader: FC<previewElementProps> = ({ content, hidden }) => (
  <p hidden={hidden} className="border-b-4 my-2 text-lg font-semibold p-2">
    {typeof content !== "string" ? <>Placeholder Text</> : <>{content}</>}
  </p>
);

const DefaultFooter: FC<previewElementProps> = ({ content, hidden }) => (
  <p hidden={hidden} className="border-t-4 my-2 text-lg font-semibold p-2">
    {typeof content !== "string" ? <>Placeholder Text</> : <>{content}</>}
  </p>
);

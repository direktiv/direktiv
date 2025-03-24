import {
  CircleCheck,
  CircleX,
  Loader2,
  Plus,
  Save,
  Settings,
  Unplug,
} from "lucide-react";
import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { FormEvent, useState } from "react";
import { KeyWithDepth, extractKeysWithDepth } from "../../utils";
import {
  LayoutSchemaType,
  PageElementContentSchema,
  PageElementContentSchemaType,
  TableContentSchemaType,
  TableKeySchema,
  TableKeySchemaType,
  TableSchemaType,
} from "~/pages/namespace/Explorer/Page/PageEditor/schema";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";

import Button from "~/design/Button";
import FilePicker from "~/components/FilePicker";
import FormErrors from "~/components/FormErrors";
import { Pagination } from "../../Pagination";
import TableColumnForm from "../utils/TableColumnForm";
import { useCreateInstanceWithOutput } from "~/api/instances/mutate/createWithOutput";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

const TableForm = ({
  layout,
  pageElementID,
  onEdit,
}: {
  layout: LayoutSchemaType;
  pageElementID: number;
  onEdit: (content: TableContentSchemaType) => void;
  onChange: (newArray: string[]) => void;
}) => {
  const nonEmptyTable = [{ header: "TableHeader 1", cell: "TableCell 1" }];
  const defaultTableData = { header: "TableHeader 1", cell: "TableCell 1" };

  const { t } = useTranslation();
  const [testSucceeded, setTestSucceeded] = useState<boolean | null>(null);
  const [selectRoute, setSelectRoute] = useState<string>("");
  const [selectItems, setSelectItems] = useState<KeyWithDepth[]>([]);
  const [tableHeaderAndCells, setTableHeaderAndCells] =
    useState<TableSchemaType>(
      layout[pageElementID]?.content?.content ?? nonEmptyTable
    );
  const [index, setIndex] = useState<number>(0);
  const [isPending, setIsPending] = useState<boolean>(false);

  const [currentColumnValue, setCurrentColumnValue] =
    useState<TableKeySchemaType>(
      tableHeaderAndCells[index] ?? defaultTableData
    );

  const formId = "edit-table-element";
  const {
    setValue,
    getValues,
    formState: { errors },
  } = useForm<PageElementContentSchemaType>({
    resolver: zodResolver(PageElementContentSchema),
    defaultValues: { content: tableHeaderAndCells },
  });

  const { mutate: getOutput } = useCreateInstanceWithOutput({
    onError: () => {
      setIsPending(false);
      setTestSucceeded(false);
    },
    onSuccess: (namespace, data) => {
      setIsPending(false);
      setTestSucceeded(true);
      setSelectItems(extractKeysWithDepth(data));
    },
  });

  const loadTableColumnIntoDisplay = (index: number) => {
    const loadedValue = tableHeaderAndCells[index] ?? defaultTableData;
    setCurrentColumnValue(loadedValue);
  };

  const updateTableColumn = (
    fieldName: "header" | "cell",
    fieldValue: string
  ) => {
    fieldName === "header"
      ? setValue(`content.${index}.header`, fieldValue)
      : setValue(`content.${index}.cell`, fieldValue);

    const headerValue = getValues(`content.${index}.header`);
    const cellValue = getValues(`content.${index}.cell`);

    const valueAtIndex = {
      header: String(headerValue) ?? "",
      cell: String(cellValue) ?? "",
    };

    const newValue =
      typeof valueAtIndex === typeof TableKeySchema
        ? valueAtIndex
        : defaultTableData;

    const prev = tableHeaderAndCells;

    if (prev.length > 1) {
      prev.splice(index, 1, newValue);
      setTableHeaderAndCells(prev);
    } else {
      setTableHeaderAndCells([defaultTableData]);
    }
  };

  const addTableColumn = () => {
    const newElement = {
      header: `Example Header ${tableHeaderAndCells.length + 1}`,
      cell: "unset",
    };
    setTableHeaderAndCells((prev) => [...prev, newElement]);
    setIndex(tableHeaderAndCells.length - 1);
  };

  const deleteTableColumn = () => {
    const prev = tableHeaderAndCells;
    if (prev.length > 1) {
      prev.splice(index, 1);
      setTableHeaderAndCells(prev);
    } else {
      setTableHeaderAndCells([defaultTableData]);
    }

    const newIndex =
      index !== 0 && index === tableHeaderAndCells.length - 1
        ? tableHeaderAndCells.length - 2
        : index;

    setIndex(newIndex);
  };

  const onSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    onEdit({ content: tableHeaderAndCells });
  };

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Settings /> Edit table component
        </DialogTitle>
      </DialogHeader>
      <FormErrors errors={errors} className="mb-5" />
      <form id={formId} onSubmit={onSubmit}>
        <div className="my-3">
          <div className="flex gap-5">
            <fieldset className="flex gap-5">
              <label className="w-[120px] text-left text-[14px]" htmlFor="text">
                Data Source:
              </label>
              <FilePicker
                onChange={setSelectRoute}
                selectable={(file) => file.type === "workflow"}
              />
            </fieldset>
            <Button
              disabled={selectRoute.length <= 1}
              variant={testSucceeded === false ? "destructive" : "outline"}
              onClick={(e) => {
                e.preventDefault();
                setIsPending(true);
                getOutput({ path: selectRoute, payload: "ha" });
              }}
            >
              {isPending && (
                <Loader2 className="animate-spin" aria-label="Loading" />
              )}
              {!isPending && testSucceeded === null && (
                <Unplug aria-label="Unplugged" />
              )}
              {!isPending && testSucceeded === true && (
                <CircleCheck aria-label="Connected" />
              )}
              {!isPending && testSucceeded === false && (
                <CircleX aria-label="Failed" />
              )}
              Connect Data Source
            </Button>
          </div>
          <div className="my-6">
            <div className="flex gap-5">
              <fieldset className="flex gap-5">
                <label
                  className="w-[120px] text-left text-[14px]"
                  htmlFor="text"
                >
                  Data Keys:
                </label>
                <TableColumnForm
                  value={currentColumnValue}
                  setValue={setCurrentColumnValue}
                  selectItems={selectItems}
                  onChange={(fieldName, fieldValue) =>
                    updateTableColumn(fieldName, fieldValue)
                  }
                  onDelete={deleteTableColumn}
                />
              </fieldset>
              <div className="w-full flex-wrap m-0">
                <div className="flex pt-4 gap-4">
                  <Pagination
                    totalPages={tableHeaderAndCells.length}
                    value={index + 1}
                    onChange={(clickedPage) => {
                      setIndex(clickedPage - 1);
                      loadTableColumnIntoDisplay(clickedPage - 1);
                    }}
                  />
                  <Button
                    icon
                    variant="outline"
                    onClick={(e) => {
                      e.preventDefault();
                      addTableColumn();
                    }}
                  >
                    <Plus />
                  </Button>
                </div>
              </div>
            </div>
          </div>

          <Table className="p-2 my-2 border-2 text-xs">
            <TableHead className="border-2">
              <TableRow className="hover:bg-transparent">
                {tableHeaderAndCells.map((data, index) => (
                  <TableHeaderCell key={index}>{data.header}</TableHeaderCell>
                ))}
              </TableRow>
            </TableHead>
            <TableBody>
              <TableRow className="border-2 hover:bg-transparent">
                {tableHeaderAndCells.map((data, index) => (
                  <TableCell key={index}>{data.cell}</TableCell>
                ))}
              </TableRow>
            </TableBody>
          </Table>
        </div>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="ghost">
              {t("pages.explorer.tree.delete.cancelBtn")}
            </Button>
          </DialogClose>
          <Button type="submit" form={formId} variant="outline">
            <Save /> Save
          </Button>
        </DialogFooter>
      </form>
    </>
  );
};

export default TableForm;

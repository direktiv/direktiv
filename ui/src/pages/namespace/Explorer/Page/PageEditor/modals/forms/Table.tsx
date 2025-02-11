import Button, { ButtonProps } from "~/design/Button";
import {
  Check,
  CircleCheck,
  CircleX,
  Loader2,
  Plus,
  Save,
  Settings,
  Trash2,
  Unplug,
  X,
} from "lucide-react";
import {
  Controller,
  SubmitHandler,
  useFieldArray,
  useForm,
  useWatch,
} from "react-hook-form";
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
  TableSchema,
  TableSchemaType,
} from "~/pages/namespace/Explorer/Page/PageEditor/schema";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";

import { Card } from "~/design/Card";
import FilePicker from "~/components/FilePicker";
import FormErrors from "~/components/FormErrors";
import Input from "~/design/Input";
import { Pagination } from "~/components/Pagination";
import { useCreateInstanceWithOutput } from "~/api/instances/mutate/createWithOutput";
import { useRoutes } from "~/api/gateway/query/getRoutes";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type fieldType = {
  id: string;
  value: {
    checkbox: "on";
  };
};

const ConditionalInput = ({
  control,
  index,
  field,
}: {
  control: any;
  index: number;
  field: fieldType;
}) => {
  const value = useWatch({
    name: "test",
    control,
  });

  return (
    <Controller
      control={control}
      name={`test.${index}.header`}
      render={({ field }) => <input {...field} />}
    />
  );
};

const TableForm = ({
  layout,
  pageElementID,
  onEdit,
  onChange,
}: {
  layout: LayoutSchemaType;
  pageElementID: number;
  onEdit: (content: unknown) => void;
  onChange: (newArray: string[]) => void;
}) => {
  const { t } = useTranslation();
  const [testSucceeded, setTestSucceeded] = useState<boolean | null>(null);
  let variant: ButtonProps["variant"] = "outline";

  const { data: routes } = useRoutes();
  const [selectRoute, setSelectRoute] = useState<string>("/ns/namespace/hd");

  const [output, setOutput] = useState<KeyWithDepth[]>([]);

  let isPending;

  type TableDataType = {
    header: string;
    cell: undefined | string;
  };

  const exampleTableData = {
    header: "Example Header",
    cell: "unset",
  };

  const exampleTableData2 = {
    header: "Example Header 2",
    cell: "unset",
  };

  const [tableHeaderAndCells, setTableHeaderAndCells] = useState([
    exampleTableData,
  ]);

  const [index, setIndex] = useState<number>(0);

  const [tableHeader, setTableHeader] = useState<string>(
    tableHeaderAndCells[index]?.header ?? ""
  );
  const [tableCell, setTableCell] = useState<string | undefined>(
    tableHeaderAndCells[index]?.cell ?? ""
  );

  const [displayTableHeader, setDisplayTableHeader] = useState<string>(
    tableHeaderAndCells[index]?.header ?? ""
  );
  const [displayTableCell, setDisplayTableCell] = useState<string | undefined>(
    tableHeaderAndCells[index]?.cell ?? ""
  );

  const onSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    e.stopPropagation(); // prevent the parent form from submitting
    //  handleSubmit(onSubmit)(e);

    onEdit(tableHeaderAndCells);
  };

  const defaultTableData = exampleTableData;

  const formId = "edit-table-element";

  const {
    register,
    handleSubmit,
    control,
    setValue,
    getValues,
    formState: { errors },
  } = useForm<PageElementContentSchemaType>({
    resolver: zodResolver(PageElementContentSchema),
    defaultValues: {
      content: { ...defaultTableData },
    },
  });

  const { fields, append, prepend } = useFieldArray({
    control,
    name: "content" as const,
  });

  if (testSucceeded === false) {
    variant = "destructive";
  }

  if (testSucceeded === true) {
    variant = "outline";
  }

  const { mutate: getOutput } = useCreateInstanceWithOutput({
    onError: () => {
      isPending = false;
      setTestSucceeded(false);
    },
    onSuccess: (namespace, data) => {
      isPending = false;
      setTestSucceeded(true);
      const outputEntries = extractKeysWithDepth(data);

      setOutput(outputEntries);
    },
  });

  const loadPaginationData = (index: number) => {
    setTableCell(tableHeaderAndCells[index]?.cell);
    setTableHeader(tableHeaderAndCells[index]?.header);
  };

  let added = false;
  const addTableItem = () => {
    added = true;
    const newPage = tableHeaderAndCells.length;

    const newcell = getValues("cell");

    const newElement = {
      header: `Example Header ${newPage + 1}`,
      cell: "unset",
    };

    setIndex(newPage);

    const copyArray = tableHeaderAndCells;
    copyArray.splice(newPage, 0, newElement);
    setTableHeaderAndCells(copyArray);

    setTableHeader(newElement.header);
    setTableCell(newElement.cell);

    setValue("header", newElement.header);
    setValue("cell", newElement.cell);
  };

  const deleteTableItem = () => {
    const actualizedIndex =
      index === tableHeaderAndCells.length - 1
        ? tableHeaderAndCells.length - 2
        : index;

    setIndex(actualizedIndex);

    const copyArray = tableHeaderAndCells;
    copyArray.splice(index, 1);
    setTableHeaderAndCells(copyArray);

    const actualizedElement =
      tableHeaderAndCells[actualizedIndex] ?? exampleTableData;

    setTableHeader(actualizedElement.header);
    setTableCell(actualizedElement.cell);
  };

  const updateTable = (element: string, value: string) => {
    if (added === true) return null;

    let newElement;

    if (element === "header") {
      newElement = {
        header: value,
        cell: tableCell,
      };
    } else {
      newElement = {
        header: tableHeader,
        cell: value,
      };
    }

    // const newElement = {
    //   header: tableHeader,
    //   cell: tableCell,
    // };

    // const newElement2 = {
    //   header: getValues("header"),
    //   cell: getValues("cell"),
    // };
    const copyArray = tableHeaderAndCells;
    copyArray.splice(index, 1, newElement);
    setTableHeaderAndCells(copyArray);
  };

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Settings /> Edit table component
        </DialogTitle>
      </DialogHeader>
      <>Index:{JSON.stringify(index)}</>

      <FormErrors errors={errors} className="mb-5" />
      <form id={formId} onSubmit={onSubmit}>
        <div className="my-3">
          <div className="flex flex-row gap-5">
            <fieldset className="flex items-start gap-5">
              <label className="w-[120px] text-left text-[14px]" htmlFor="text">
                Data Source:
              </label>
              <div className="w-full flex-wrap">
                <FilePicker
                  onChange={(value) => setSelectRoute(value)}
                  selectable={(file) => file.type === "workflow"}
                />
              </div>
            </fieldset>
            <Button
              variant={variant}
              onClick={(e) => {
                e.preventDefault();
                isPending = true;

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
            <div className="flex flex-row gap-5">
              <fieldset className="flex items-start gap-5">
                <label
                  className="w-[120px] text-left text-[14px]"
                  htmlFor="text"
                >
                  Data Keys:
                </label>

                <div className="w-full m-0">
                  <Card noShadow className="flex flex-row ">
                    <div className="flex flex-col">
                      <div className="flex flex-row">
                        <Button
                          className="w-32 rounded-none rounded-tl-md bg-gray-2"
                          variant="outline"
                          asChild
                        >
                          <label>Table Header</label>
                        </Button>

                        <Input
                          placeholder="Insert a Caption for the data below"
                          className="w-80 rounded-none rounded-tr-md"
                          value={tableHeader}
                          onChange={(e) => {
                            setTableHeader(e.target.value);
                            setValue(`content.${index}.header`, e.target.value);
                            if (e.target.value !== displayTableHeader)
                              updateTable("header", e.target.value);
                          }}
                        />
                      </div>
                      <div className="flex flex-row ">
                        <Button
                          className="w-32 rounded-none rounded-bl-md bg-gray-2"
                          variant="outline"
                          asChild
                        >
                          <label>Table Cell</label>
                        </Button>

                        <>
                          {output.length ? (
                            <Select
                              value={tableCell}
                              onValueChange={(newValue) => {
                                setTableCell(newValue);
                                setValue(`content.${index}.cell`, newValue);
                                if (newValue !== tableCell)
                                  updateTable("cell", newValue);
                              }}
                            >
                              <SelectTrigger
                                defaultValue="unset"
                                className="w-80 justify-start rounded-none rounded-br-md"
                                variant="outline"
                                id="scale"
                              >
                                <SelectValue placeholder="Select Data" />
                              </SelectTrigger>
                              <SelectContent>
                                <SelectGroup>
                                  {output?.map((element, key) => (
                                    <SelectItem
                                      key={element.key}
                                      value={element.key}
                                    >
                                      {element.key}
                                    </SelectItem>
                                  ))}
                                </SelectGroup>
                              </SelectContent>
                            </Select>
                          ) : (
                            <Select>
                              <SelectTrigger
                                className="w-80 justify-start"
                                variant="outline"
                                id="scale"
                                disabled
                              >
                                Connect a Data Source first
                              </SelectTrigger>
                            </Select>
                          )}
                        </>
                      </div>
                      <div className="flex-col flex items-end ">
                        <Button
                          className=""
                          icon
                          variant="outline"
                          onClick={() => deleteTableItem()}
                        >
                          <Trash2 />
                        </Button>
                      </div>
                    </div>

                    <div className="flex justify-items-end relative right-0"></div>
                  </Card>

                  <div className="flex flex-row pt-4 gap-4">
                    <Pagination
                      totalPages={tableHeaderAndCells.length}
                      value={index + 1}
                      onChange={(clickedPage) => {
                        loadPaginationData(clickedPage - 1);
                        setIndex(clickedPage - 1);
                      }}
                    />
                    <Button
                      icon
                      variant="outline"
                      onClick={() => addTableItem()}
                    >
                      <Plus />
                    </Button>
                  </div>
                </div>
              </fieldset>
            </div>
          </div>
          <br></br>
          {fields.map((field, index) => {
            const id = `test.${index}.checkbox`;
            return (
              <div key={id}>
                <section>
                  <label htmlFor={id}>Show Input</label>
                  <input
                    type="checkbox"
                    value="on"
                    id={id}
                    {...register(`content.${index}`)}
                  />
                  <ConditionalInput {...{ control, index, field }} />
                </section>
                <hr />
              </div>
            );
          })}

          <br></br>
          <button
            type="button"
            onClick={() =>
              append({
                header: "append value",
                cell: "unset",
              })
            }
          >
            append
          </button>

          <button
            type="button"
            onClick={() =>
              prepend({
                firstName: "prepend value",
              })
            }
          >
            prepend
          </button>
          <br></br>

          <Controller
            control={control}
            name={`content.${index}.header`}
            render={({ field }) => (
              <Input
                placeholder="Insert a Caption for the data below"
                className="hidden w-80 rounded-none rounded-tr-md"
                value={tableHeader}
                onChange={(e) => {
                  setTableHeader(e.target.value);
                  setValue(`content.${index}.header`, e.target.value);
                }}
              />
            )}
          />
          <Controller
            control={control}
            name={`content.${index}.cell`}
            render={({ field }) => (
              <Input
                placeholder="Insert a Caption for the data below"
                className="hidden w-80 rounded-none rounded-tr-md"
                value={tableCell}
                onChange={(e) => {
                  setTableCell(e.target.value);
                  setValue(`content.${index}.cell`, e.target.value);
                }}
              />
            )}
          />
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
          <Button type="submit" form={formId ?? undefined} variant="outline">
            <Save />
            Save
          </Button>
        </DialogFooter>
      </form>
    </>
  );
};

export default TableForm;

import { Controller, useForm } from "react-hook-form";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";
import { TableKeySchema, TableKeySchemaType } from "../../schema";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Input from "~/design/Input";
import { KeyWithDepth } from "../../utils";
import { Trash2 } from "lucide-react";
import { useEffect } from "react";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

const TableColumnForm = ({
  value,
  selectItems,
  onChange,
  onDelete,
}: {
  value: TableKeySchemaType;
  setValue: (val: TableKeySchemaType) => void;
  selectItems: KeyWithDepth[];
  onChange: (fieldName: "header" | "cell", fieldValue: string) => void;
  onDelete: () => void;
}) => {
  const { t } = useTranslation();
  const { control, reset } = useForm<TableKeySchemaType>({
    resolver: zodResolver(TableKeySchema),
    values: value,
    defaultValues: { header: "testheader", cell: "testcell" },
  });

  useEffect(() => {
    reset({
      header: value.header,
      cell: value.cell,
    });
  }, [value, reset]);

  return (
    <>
      <Card noShadow className="flex flex-row ">
        <div className="flex flex-col">
          <div className="flex flex-row">
            <Button
              className="w-32 rounded-none rounded-tl-md bg-gray-2"
              variant="outline"
              asChild
            >
              <label>
                {t(
                  "pages.explorer.page.editor.form.modals.edit.table.tableColumnForm.labelOne"
                )}
              </label>
            </Button>
            <Controller
              control={control}
              name="header"
              render={({ field }) => (
                <Input
                  value={field.value}
                  onChange={(e) => {
                    field.onChange(e.target.value);
                    onChange("header", e.target.value);
                  }}
                  placeholder={t(
                    "pages.explorer.page.editor.form.modals.edit.table.tableColumnForm.placeholderOne"
                  )}
                  className="w-80 rounded-none rounded-tr-md"
                />
              )}
            />
          </div>
          <div className="flex flex-row ">
            <Button
              className="w-32 rounded-none rounded-bl-md bg-gray-2"
              variant="outline"
              asChild
            >
              <label>
                {t(
                  "pages.explorer.page.editor.form.modals.edit.table.tableColumnForm.labelTwo"
                )}
              </label>
            </Button>
            <>
              {selectItems.length ? (
                <Controller
                  control={control}
                  name="cell"
                  render={({ field }) => (
                    <Select
                      value={field.value}
                      onValueChange={(newValue) => {
                        field.onChange(newValue);
                        onChange("cell", newValue);
                      }}
                    >
                      <SelectTrigger
                        defaultValue="unset"
                        className="w-80 justify-start rounded-none rounded-br-md"
                        variant="outline"
                        id="scale"
                      >
                        <SelectValue
                          placeholder={t(
                            "pages.explorer.page.editor.form.modals.edit.table.tableColumnForm.placeholderTwo"
                          )}
                        />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectGroup>
                          {selectItems?.map((element) => (
                            <SelectItem key={element.key} value={element.key}>
                              {element.key}
                            </SelectItem>
                          ))}
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                  )}
                />
              ) : (
                <Select>
                  <SelectTrigger
                    className="w-80 justify-start"
                    variant="outline"
                    id="scale"
                    disabled
                  >
                    {t(
                      "pages.explorer.page.editor.form.modals.edit.table.tableColumnForm.selectBtn"
                    )}
                  </SelectTrigger>
                </Select>
              )}
            </>
          </div>
          <div className="flex-col flex items-end m-0">
            <Button
              className="m-0"
              icon
              variant="outline"
              onClick={(e) => {
                e.preventDefault();
                onDelete();
              }}
            >
              <Trash2 />
            </Button>
          </div>
        </div>
      </Card>
    </>
  );
};

export default TableColumnForm;

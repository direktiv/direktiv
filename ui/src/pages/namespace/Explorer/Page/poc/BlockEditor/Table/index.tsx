import { Table as TableSchema, TableType } from "../../schema/blocks/table";

import { BlockEditFormProps } from "..";
import { ColumnForm } from "./ColumnForm";
import { DialogFooter } from "../components/Footer";
import { DialogHeader } from "../components/Header";
import { Table as TableForm } from "../components/FormElements/Table";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type TableEditFormProps = BlockEditFormProps<TableType>;

export const Table = ({
  action,
  block: propBlock,
  path,
  onSubmit,
}: TableEditFormProps) => {
  const { t } = useTranslation();
  const form = useForm<TableType>({
    resolver: zodResolver(TableSchema),
    defaultValues: propBlock,
  });

  return (
    <>
      <DialogHeader action={action} path={path} type={propBlock.type} />
      {/* TODO: connect form and display errors */}
      <div className="text-gray-10 dark:text-gray-10">
        {t("direktivPage.blockEditor.blockForms.table.description")}
      </div>
      {/* TODO: loop.id and loop.data */}
      {/* TODO: actions */}
      <TableForm
        data={form.getValues("columns")}
        onChange={(newValue) => {
          form.setValue("columns", newValue);
        }}
        itemLabel={t(
          "direktivPage.blockEditor.blockForms.table.column.itemLabel"
        )}
        label={(count) =>
          t("direktivPage.blockEditor.blockForms.table.column.tableLabel", {
            count,
          })
        }
        renderRow={(query) => [query.label, query.content]}
        getItemKey={(query, index) => index}
        renderForm={(formId, onSubmit, defaultValues) => (
          <ColumnForm
            formId={formId}
            onSubmit={onSubmit}
            defaultValues={defaultValues}
          />
        )}
      />
      <DialogFooter onSubmit={() => onSubmit(form.getValues())} />
    </>
  );
};

import { Table as TableSchema, TableType } from "../../schema/blocks/table";

import { ActionForm } from "./ActionForm";
import { BlockEditFormProps } from "..";
import { ColumnForm } from "./ColumnForm";
import { Fieldset } from "~/components/Form/Fieldset";
import { FormWrapper } from "../components/FormWrapper";
import Input from "~/design/Input";
import { Table as TableForm } from "../components/FormElements/Table";
import { VariablePicker } from "../components/VariablePicker";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type TableEditFormProps = BlockEditFormProps<TableType>;

export const Table = ({
  action,
  block: propBlock,
  path,
  onSubmit,
  onCancel,
}: TableEditFormProps) => {
  const { t } = useTranslation();
  const form = useForm<TableType>({
    resolver: zodResolver(TableSchema),
    defaultValues: propBlock,
  });

  return (
    <FormWrapper
      description={t("direktivPage.blockEditor.blockForms.table.description")}
      form={form}
      block={propBlock}
      action={action}
      path={path}
      onSubmit={onSubmit}
      onCancel={onCancel}
    >
      <Fieldset
        label={t("direktivPage.blockEditor.blockForms.table.data.idLabel")}
        htmlFor="data-id"
      >
        <Input
          {...form.register("data.id")}
          id="data-id"
          placeholder={t(
            "direktivPage.blockEditor.blockForms.table.data.idPlaceholder"
          )}
        />
      </Fieldset>
      <Fieldset
        label={t("direktivPage.blockEditor.blockForms.table.data.dataLabel")}
        htmlFor="data-data"
      >
        <VariablePicker
          value={form.watch("data.data")}
          onUpdate={(value) => form.setValue("data.data", value)}
          id="data-data"
          placeholder={t(
            "direktivPage.blockEditor.blockForms.table.data.dataPlaceholder"
          )}
        />
      </Fieldset>
      <TableForm
        data={form.getValues("actions")}
        onChange={(newValue) => {
          form.setValue("actions", newValue);
        }}
        itemLabel={t(
          "direktivPage.blockEditor.blockForms.table.action.itemLabel"
        )}
        label={(count) =>
          t("direktivPage.blockEditor.blockForms.table.action.tableLabel", {
            count,
          })
        }
        renderRow={(query) => [query.label]}
        renderForm={(formId, onSubmit, defaultValues) => (
          <ActionForm
            formId={formId}
            onSubmit={onSubmit}
            defaultValues={defaultValues}
          />
        )}
      />
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
        renderForm={(formId, onSubmit, defaultValues) => (
          <ColumnForm
            formId={formId}
            onSubmit={onSubmit}
            defaultValues={defaultValues}
          />
        )}
      />
      <Fieldset
        label={t("direktivPage.blockEditor.blockForms.loop.pageSizeLabel")}
        htmlFor="data-pageSize"
      >
        <Input
          {...form.register("data.pageSize", { valueAsNumber: true })}
          id="data-pageSize"
          type="number"
          placeholder={t(
            "direktivPage.blockEditor.blockForms.loop.pageSizePlaceholder"
          )}
        />
      </Fieldset>
    </FormWrapper>
  );
};

import { Table as TableSchema, TableType } from "../../schema/blocks/table";

import { ActionForm } from "./ActionForm";
import { BlockEditFormProps } from "..";
import { ColumnForm } from "./ColumnForm";
import { Fieldset } from "~/components/Form/Fieldset";
import { FormWrapper } from "../components/FormWrapper";
import Input from "~/design/Input";
import { Table as TableForm } from "../components/FormElements/Table";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type TableEditFormProps = BlockEditFormProps<TableType>;

export const Table = ({ block: propBlock, onSubmit }: TableEditFormProps) => {
  const { t } = useTranslation();
  const form = useForm<TableType>({
    resolver: zodResolver(TableSchema),
    defaultValues: propBlock,
  });

  return (
    <FormWrapper form={form} onSubmit={onSubmit}>
      <div className="text-gray-10 dark:text-gray-10">
        {t("direktivPage.blockEditor.blockForms.table.description")}
      </div>
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
        <Input
          {...form.register("data.data")}
          id="id"
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
    </FormWrapper>
  );
};

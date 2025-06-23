import { GenericTable } from "./GenericTable";
import { QueryForm } from "./QueryForm";
import { QueryType } from "../../schema/procedures/query";
import { useTranslation } from "react-i18next";

type QueriesTableProps = {
  defaultValue: QueryType[];
  onChange: (newValues: QueryType[]) => void;
};

export const QueriesTable = ({ defaultValue, onChange }: QueriesTableProps) => {
  const { t } = useTranslation();
  return (
    <GenericTable
      data={defaultValue}
      onChange={onChange}
      label={t("direktivPage.blockEditor.blockForms.queryProvider.queryLabel")}
      renderRow={(query) => [query.id, query.url]}
      getItemKey={(query) => query.id}
      renderForm={(formId, onSubmit, defaultValues) => (
        <QueryForm
          formId={formId}
          onSubmit={onSubmit}
          defaultValues={defaultValues}
        />
      )}
    />
  );
};

import { GenericTable } from "./GenericTable";
import { QueryForm } from "./QueryForm";
import { QueryType } from "../../schema/procedures/query";

type QueriesTableProps = {
  defaultValue: QueryType[];
  onChange: (newValues: QueryType[]) => void;
};

export const QueriesTable = ({ defaultValue, onChange }: QueriesTableProps) => (
  <GenericTable
    data={defaultValue}
    onChange={onChange}
    itemName="Query"
    itemNamePlural="Queries"
    renderRow={(query) => [query.id, query.url]}
    getItemKey={(query) => query.id}
    formTitle="Query"
    renderForm={(formId, onSubmit, defaultValues) => (
      <QueryForm
        formId={formId}
        onSubmit={onSubmit}
        defaultValues={defaultValues}
      />
    )}
    addButtonText="add Query"
  />
);

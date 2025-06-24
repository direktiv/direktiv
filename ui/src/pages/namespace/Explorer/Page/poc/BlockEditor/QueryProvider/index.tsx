import {
  QueryProvider as QueryProviderSchema,
  QueryProviderType,
} from "../../schema/blocks/queryProvider";

import { BlockEditFormProps } from "..";
import { DialogFooter } from "../components/Footer";
import { DialogHeader } from "../components/Header";
import { QueryForm } from "./QueryForm";
import { Table } from "../components/FormElements/Table";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type HeadlineEditFormProps = BlockEditFormProps<QueryProviderType>;

export const QueryProvider = ({
  action,
  block: propBlock,
  path,
  onSubmit,
}: HeadlineEditFormProps) => {
  const { t } = useTranslation();
  const form = useForm<QueryProviderType>({
    resolver: zodResolver(QueryProviderSchema),
    defaultValues: propBlock,
  });

  return (
    <>
      <DialogHeader action={action} path={path} type={propBlock.type} />
      <Table
        data={form.getValues("queries")}
        onChange={(newValue) => {
          form.setValue("queries", newValue);
        }}
        label={t(
          "direktivPage.blockEditor.blockForms.queryProvider.queryLabel"
        )}
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
      <DialogFooter onSubmit={() => onSubmit(form.getValues())} />
    </>
  );
};

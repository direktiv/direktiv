import {
  QueryProvider as QueryProviderSchema,
  QueryProviderType,
} from "../../schema/blocks/queryProvider";

import { BlockEditFormProps } from "..";
import { Footer } from "../components/Footer";
import { Header } from "../components/Header";
import { QueryForm } from "./QueryForm";
import { Table } from "../components/FormElements/Table";
import { queryToUrl } from "../utils";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type QueryProviderEditFormProps = BlockEditFormProps<QueryProviderType>;

export const QueryProvider = ({
  action,
  block: propBlock,
  path,
  onSubmit,
  onCancel,
}: QueryProviderEditFormProps) => {
  const { t } = useTranslation();
  const form = useForm<QueryProviderType>({
    resolver: zodResolver(QueryProviderSchema),
    defaultValues: propBlock,
  });

  return (
    <div className="flex flex-col gap-4 p-1">
      <Header action={action} path={path} block={propBlock} />
      <div className="text-gray-10">
        {t("direktivPage.blockEditor.blockForms.queryProvider.description")}
      </div>
      <Table
        data={form.getValues("queries")}
        onChange={(newValue) => {
          form.setValue("queries", newValue);
        }}
        itemLabel={t(
          "direktivPage.blockEditor.blockForms.queryProvider.query.itemLabel"
        )}
        label={(count) =>
          t(
            "direktivPage.blockEditor.blockForms.queryProvider.query.tableLabel",
            { count }
          )
        }
        renderRow={(query) => [query.id, queryToUrl(query)]}
        getItemKey={(query) => query.id}
        renderForm={(formId, onSubmit, defaultValues) => (
          <QueryForm
            formId={formId}
            onSubmit={onSubmit}
            defaultValues={defaultValues}
          />
        )}
      />
      <Footer
        onSubmit={() => onSubmit(form.getValues())}
        onCancel={() => onCancel}
      />
    </div>
  );
};

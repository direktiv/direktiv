import {
  QueryProvider as QueryProviderSchema,
  QueryProviderType,
} from "../../schema/blocks/queryProvider";

import { BlockEditFormProps } from "..";
import { DialogFooter } from "../components/Footer";
import { DialogHeader } from "../components/Header";
import { QueryForm } from "./QueryForm";
import { Table } from "../components/FormElements/Table";
import { keyValueArrayToObject } from "../../PageCompiler/primitives/keyValue/utils";
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
        renderRow={(query) => {
          let url = query.url;
          const searchParams = new URLSearchParams(
            keyValueArrayToObject(query.queryParams ?? [])
          );
          const queryString = searchParams.toString();
          if (queryString) {
            url = url.concat("?", queryString);
          }
          return [query.id, url];
        }}
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

import {
  QueryProvider as QueryProviderSchema,
  QueryProviderType,
} from "../../schema/blocks/queryProvider";

import { BlockEditFormProps } from "..";
import { FormWrapper } from "../components/FormWrapper";
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
}: QueryProviderEditFormProps) => {
  const { t } = useTranslation();
  const form = useForm<QueryProviderType>({
    resolver: zodResolver(QueryProviderSchema),
    defaultValues: propBlock,
  });

  return (
    <FormWrapper
      description={t(
        "direktivPage.blockEditor.blockForms.queryProvider.description"
      )}
      form={form}
      onSubmit={onSubmit}
      action={action}
      path={path}
      blockType={propBlock.type}
    >
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
        renderForm={(formId, onSubmit, defaultValues) => (
          <QueryForm
            formId={formId}
            onSubmit={onSubmit}
            defaultValues={defaultValues}
          />
        )}
      />
    </FormWrapper>
  );
};

import FormErrors, { errorsType } from "~/components/FormErrors";
import {
  TableColumn,
  TableColumnType,
} from "../../schema/blocks/table/tableColumn";

import { Fieldset } from "~/components/Form/Fieldset";
import Input from "~/design/Input";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type ColumnFormProps = {
  defaultValues?: TableColumnType;
  formId: string;
  onSubmit: (data: TableColumnType) => void;
};

export const ColumnForm = ({
  defaultValues,
  formId,
  onSubmit,
}: ColumnFormProps) => {
  const { t } = useTranslation();
  const {
    handleSubmit,
    register,
    formState: { errors },
  } = useForm<TableColumnType>({
    resolver: zodResolver(TableColumn),
    defaultValues,
  });

  return (
    <form onSubmit={handleSubmit(onSubmit)} id={formId}>
      {errors && <FormErrors errors={errors as errorsType} className="mb-5" />}
      <Fieldset
        label={t("direktivPage.blockEditor.blockForms.table.column.labelLabel")}
        htmlFor="label"
      >
        <Input
          {...register("label")}
          id="label"
          placeholder={t(
            "direktivPage.blockEditor.blockForms.table.column.labelPlaceholder"
          )}
        />
      </Fieldset>
      <Fieldset
        label={t(
          "direktivPage.blockEditor.blockForms.table.column.contentLabel"
        )}
        htmlFor="content"
      >
        <Input
          {...register("content")}
          id="content"
          placeholder={t(
            "direktivPage.blockEditor.blockForms.table.column.contentPlaceholder"
          )}
        />
      </Fieldset>
    </form>
  );
};

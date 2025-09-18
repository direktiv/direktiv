import FormErrors, { errorsType } from "~/components/FormErrors";
import {
  TableColumn,
  TableColumnType,
} from "../../schema/blocks/table/tableColumn";

import { Fieldset } from "~/components/Form/Fieldset";
import Input from "~/design/Input";
import { SmartInput } from "../components/SmartInput";
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
    formState: { errors },
    handleSubmit,
    register,
    setValue,
    watch,
  } = useForm<TableColumnType>({
    resolver: zodResolver(TableColumn),
    defaultValues: {
      type: "table-column",
      content: "",
      label: "",
      ...defaultValues,
    },
  });

  const onFormSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.stopPropagation();
    handleSubmit(onSubmit)(e);
  };

  return (
    <form onSubmit={onFormSubmit} id={formId}>
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
        <SmartInput
          value={watch("content")}
          onUpdate={(value) => setValue("content", value)}
          id="content"
          placeholder={t(
            "direktivPage.blockEditor.blockForms.table.column.contentPlaceholder"
          )}
        />
      </Fieldset>
    </form>
  );
};

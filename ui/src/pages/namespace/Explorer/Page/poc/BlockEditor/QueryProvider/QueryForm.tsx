import { Controller, useForm } from "react-hook-form";
import FormErrors, { errorsType } from "~/components/FormErrors";
import { Query, QueryType } from "../../schema/procedures/query";

import { Fieldset } from "~/components/Form/Fieldset";
import Input from "~/design/Input";
import { KeyValueInput } from "../components/FormElements/KeyValueInput";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type QueryFormProps = {
  defaultValues?: QueryType;
  formId: string;
  onSubmit: (data: QueryType) => void;
};

export const QueryForm = ({
  defaultValues,
  formId,
  onSubmit,
}: QueryFormProps) => {
  const { t } = useTranslation();
  const {
    handleSubmit,
    register,
    control,
    formState: { errors },
  } = useForm<QueryType>({
    resolver: zodResolver(Query),
    defaultValues: {
      id: "",
      url: "",
      queryParams: [],
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
        label={t(
          "direktivPage.blockEditor.blockForms.queryProvider.query.idLabel"
        )}
        htmlFor="id"
      >
        <Input
          {...register("id")}
          id="id"
          placeholder={t(
            "direktivPage.blockEditor.blockForms.queryProvider.query.idPlaceholder"
          )}
        />
      </Fieldset>
      <Fieldset
        label={t(
          "direktivPage.blockEditor.blockForms.queryProvider.query.urlLabel"
        )}
        htmlFor="url"
      >
        <Input
          {...register("url")}
          id="url"
          placeholder={t(
            "direktivPage.blockEditor.blockForms.queryProvider.query.urlPlaceholder"
          )}
        />
      </Fieldset>
      <Controller
        control={control}
        name="queryParams"
        render={({ field }) => (
          <KeyValueInput
            field={field}
            label={t(
              "direktivPage.blockEditor.blockForms.queryProvider.query.queryParamsLabel"
            )}
          />
        )}
      />
    </form>
  );
};

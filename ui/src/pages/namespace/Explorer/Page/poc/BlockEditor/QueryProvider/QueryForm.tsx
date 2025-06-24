import { Controller, useForm } from "react-hook-form";
import FormErrors, { errorsType } from "~/components/FormErrors";
import { Query, QueryType } from "../../schema/procedures/query";

import { Fieldset } from "~/components/Form/Fieldset";
import { FormEvent } from "react";
import Input from "~/design/Input";
import { KeyValueInput } from "../components/FormElements/QueryParamsForm";
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
  const {
    handleSubmit,
    register,
    control,
    formState: { errors },
  } = useForm<QueryType>({
    resolver: zodResolver(Query),
    defaultValues,
  });

  const submitForm = (e: FormEvent<HTMLFormElement>) => {
    e.stopPropagation(); // prevent the parent form from submitting
    handleSubmit(onSubmit)(e);
  };

  // TODO: i18n
  return (
    <form onSubmit={submitForm} id={formId}>
      {errors && <FormErrors errors={errors as errorsType} className="mb-5" />}
      <Fieldset label="id" htmlFor="id">
        <Input {...register("id")} id="id" />
      </Fieldset>
      <Fieldset label="url" htmlFor="url">
        <Input {...register("url")} id="url" />
      </Fieldset>
      <Controller
        control={control}
        name="queryParams"
        render={({ field }) => (
          <KeyValueInput field={field} label="Query Parameters" />
        )}
      />
    </form>
  );
};

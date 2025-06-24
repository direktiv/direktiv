import { Controller, useForm } from "react-hook-form";
import FormErrors, { errorsType } from "~/components/FormErrors";
import { Query, QueryType } from "../../schema/procedures/query";

import { ArrayForm } from "~/components/Form/Array";
import { Fieldset } from "~/components/Form/Fieldset";
import { FormEvent } from "react";
import Input from "~/design/Input";
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
      <Fieldset label="Query Parameters">
        <Controller
          control={control}
          name="queryParams"
          render={({ field }) => (
            <ArrayForm
              defaultValue={field.value || []}
              onChange={field.onChange}
              emptyItem={{ key: "", value: "" }}
              itemIsValid={(item) =>
                !!item && Object.values(item).every((v) => v !== "")
              }
              renderItem={({ value: objectValue, setValue, handleKeyDown }) => (
                <>
                  {Object.entries(objectValue).map(([key, value]) => {
                    const typedKey = key as keyof typeof objectValue;
                    return (
                      <Input
                        key={key}
                        placeholder={
                          typedKey === "key"
                            ? "Parameter name"
                            : "Parameter value"
                        }
                        value={value}
                        onKeyDown={handleKeyDown}
                        onChange={(e) => {
                          const newObject = {
                            ...objectValue,
                            [key]: e.target.value,
                          };
                          setValue(newObject);
                        }}
                      />
                    );
                  })}
                </>
              )}
            />
          )}
        />
      </Fieldset>
    </form>
  );
};

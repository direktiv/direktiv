import { Controller, SubmitHandler, useForm } from "react-hook-form";
import {
  SecretFormCreateEditSchema,
  SecretFormCreateEditSchemaType,
} from "~/api/secrets/schema";
import { decode, encode } from "js-base64";

import { DialogHeader } from "~/design/Dialog";
import FileUpload from "~/components/VariableForm/FileUpload";
import FormErrors from "~/components/FormErrors";
import Input from "~/design/Input";
import { Textarea } from "~/design/TextArea";
import { isMimeTypeEditable } from "~/components/VariableForm/utils";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type SecretFormProps = {
  defaultValues: SecretFormCreateEditSchemaType;
  dialogTitle: JSX.Element;
  dialogFooter: JSX.Element;
  unallowedNames?: string[];
  onMutate: (data: SecretFormCreateEditSchemaType) => void;
};

export const SecretForm = ({
  defaultValues,
  dialogTitle,
  dialogFooter,
  unallowedNames,
  onMutate,
}: SecretFormProps) => {
  const { t } = useTranslation();
  const {
    register,
    handleSubmit,
    control,
    setValue,
    setError,
    formState: { errors },
  } = useForm<SecretFormCreateEditSchemaType>({
    resolver: zodResolver(
      SecretFormCreateEditSchema.refine(
        (fields) =>
          !(unallowedNames ?? []).some(
            (unallowedName) => unallowedName === fields.name
          ),
        {
          // TODO:
          message: t("components.variableForm.name.nameAlreadyExists"),
        }
      )
    ),
    defaultValues,
  });

  const onSubmit: SubmitHandler<SecretFormCreateEditSchemaType> = (data) => {
    onMutate(data);
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col space-y-5">
      <DialogHeader>
        <DialogHeader>{dialogTitle}</DialogHeader>
      </DialogHeader>
      <FormErrors errors={errors} className="mb-5" />
      <fieldset className="flex items-center gap-5">
        <label className="w-[150px] text-right" htmlFor="name">
          {/* TODO: */}
          {t("components.variableForm.name.label")}
        </label>

        {/* TODO: */}
        <Input
          id="name"
          {...register("name")}
          placeholder={t("components.variableForm.name.placeholder")}
        />
      </fieldset>
      <FileUpload
        onChange={({ base64String, mimeType }) => {
          const isSupported = isMimeTypeEditable(mimeType);
          if (!isSupported) {
            setError("data", {
              message: t("pages.settings.secrets.create.unsupported"),
            });
            return;
          }
          setValue("data", base64String);
        }}
      />
      <fieldset className="flex items-start gap-5">
        <Controller
          control={control}
          name="data"
          render={({ field }) => (
            <Textarea
              defaultValue={decode(field.value)}
              className="h-96"
              onChange={(e) => {
                if (e.target.value) {
                  field.onChange(encode(e.target.value));
                }
              }}
            />
          )}
        />
      </fieldset>
      {dialogFooter}
    </form>
  );
};

export default SecretFormProps;

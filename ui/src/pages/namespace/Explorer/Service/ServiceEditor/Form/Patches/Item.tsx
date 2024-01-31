import { Controller, useForm } from "react-hook-form";
import { FC, FormEvent } from "react";
import {
  PatchOperationType,
  PatchOperations,
  PatchSchema,
  PatchSchemaType,
} from "../../schema";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import { Fieldset } from "~/components/Form/Fieldset";
import FormErrors from "~/components/FormErrors";
import Input from "~/design/Input";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type PatchItemFormProps = {
  value: PatchSchemaType | undefined;
  onSubmit: (item: PatchSchemaType) => void;
  formId: string;
};

export const PatchItemForm: FC<PatchItemFormProps> = ({
  value,
  formId,
  onSubmit,
}) => {
  const { t } = useTranslation();
  const theme = useTheme();

  const {
    handleSubmit,
    formState: { errors },
    setValue,
    register,
    watch,
    control,
  } = useForm<PatchSchemaType>({
    resolver: zodResolver(PatchSchema),
    defaultValues: {
      ...value,
    },
  });

  const submitForm = (event: FormEvent<HTMLFormElement>) => {
    event.stopPropagation(); // prevent the parent form from submitting
    handleSubmit(onSubmit)(event);
  };

  return (
    <form onSubmit={submitForm} id={formId}>
      <FormErrors errors={errors} className="mb-5" />
      <Fieldset
        label={t("pages.explorer.service.editor.form.patches.modal.op.label")}
        htmlFor="op"
      >
        <Select
          value={watch("op")}
          onValueChange={(value) => {
            setValue("op", value as PatchOperationType);
          }}
        >
          <SelectTrigger id="op" variant="outline">
            <SelectValue
              placeholder={t(
                "pages.explorer.service.editor.form.patches.modal.op.placeholder"
              )}
            />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              {PatchOperations.map((op) => (
                <SelectItem key={op} value={op}>
                  {op}
                </SelectItem>
              ))}
            </SelectGroup>
          </SelectContent>
        </Select>
      </Fieldset>
      <Fieldset
        label={t("pages.explorer.service.editor.form.patches.modal.path")}
        htmlFor="path"
      >
        <Input type="text" id="path" {...register("path")} />
      </Fieldset>
      <Fieldset
        label={t("pages.explorer.service.editor.form.patches.modal.value")}
      >
        <Card className="h-[200px] w-full p-5" background="weight-1" noShadow>
          <Controller
            control={control}
            name="value"
            render={({ field }) => (
              <Editor
                theme={theme ?? undefined}
                language="javascript"
                value={field.value}
                onChange={field.onChange}
              />
            )}
          />
        </Card>
      </Fieldset>
    </form>
  );
};

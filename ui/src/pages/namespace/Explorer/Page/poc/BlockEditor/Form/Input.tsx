import { Controller, useForm } from "react-hook-form";
import {
  FormInput as FormInputSchema,
  FormInputType,
  inputTypes,
} from "../../schema/blocks/form/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import { BaseForm } from "./BaseForm";
import { BlockEditFormProps } from "..";
import { Fieldset } from "~/components/Form/Fieldset";
import { FormWrapper } from "../components/FormWrapper";
import InputDesignComponent from "~/design/Input";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type InputProps = BlockEditFormProps<FormInputType>;

export const Input = ({
  action,
  block: propBlock,
  path,
  onSubmit,
  onCancel,
}: InputProps) => {
  const { t } = useTranslation();
  const form = useForm<FormInputType>({
    resolver: zodResolver(FormInputSchema),
    defaultValues: propBlock,
  });

  return (
    <FormWrapper
      description={t(
        "direktivPage.blockEditor.blockForms.formPrimitives.input.description"
      )}
      form={form}
      block={propBlock}
      action={action}
      path={path}
      onSubmit={onSubmit}
      onCancel={onCancel}
    >
      <BaseForm form={form} />
      <Fieldset
        label={t(
          "direktivPage.blockEditor.blockForms.formPrimitives.input.variantLabel"
        )}
        htmlFor="variant"
      >
        <Controller
          control={form.control}
          name="variant"
          render={({ field }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger variant="outline" id="variant">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {inputTypes.map((item) => (
                  <SelectItem key={item} value={item}>
                    <span>{item}</span>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        />
      </Fieldset>
      <Fieldset
        label={t(
          "direktivPage.blockEditor.blockForms.formPrimitives.input.defaultValueLabel"
        )}
        htmlFor="defaultValue"
      >
        <InputDesignComponent
          {...form.register("defaultValue")}
          id="defaultValue"
          placeholder={t(
            "direktivPage.blockEditor.blockForms.formPrimitives.input.defaultValuePlaceholder"
          )}
        />
      </Fieldset>
    </FormWrapper>
  );
};

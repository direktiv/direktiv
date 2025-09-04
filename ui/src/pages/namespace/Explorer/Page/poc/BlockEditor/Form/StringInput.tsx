import { Controller, useForm } from "react-hook-form";
import {
  FormStringInput,
  FormStringInputType,
  stringInputTypes,
} from "../../schema/blocks/form/stringInput";
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
import { SmartInput } from "../components/SmartInput";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type StringInputProps = BlockEditFormProps<FormStringInputType>;

export const StringInput = ({
  action,
  block: propBlock,
  path,
  onSubmit,
  onCancel,
}: StringInputProps) => {
  const { t } = useTranslation();
  const form = useForm<FormStringInputType>({
    resolver: zodResolver(FormStringInput),
    defaultValues: propBlock,
  });

  return (
    <FormWrapper
      description={t(
        "direktivPage.blockEditor.blockForms.formPrimitives.stringInput.description"
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
          "direktivPage.blockEditor.blockForms.formPrimitives.stringInput.variantLabel"
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
                {stringInputTypes.map((item) => (
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
          "direktivPage.blockEditor.blockForms.formPrimitives.defaultValue.label"
        )}
        htmlFor="defaultValue"
      >
        <SmartInput
          value={form.watch("defaultValue")}
          onChange={(content) => form.setValue("defaultValue", content)}
          id="defaultValue"
          placeholder={t(
            "direktivPage.blockEditor.blockForms.formPrimitives.defaultValue.placeholder"
          )}
        />
      </Fieldset>
    </FormWrapper>
  );
};

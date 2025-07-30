import { Controller, UseFormReturn, useForm } from "react-hook-form";
import {
  FormSelect as FormSelectSchema,
  FormSelectType,
} from "../../schema/blocks/form/select";

import { ArrayForm } from "~/components/Form/Array";
import { BaseForm } from "./BaseForm";
import { BlockEditFormProps } from "..";
import { Fieldset } from "~/components/Form/Fieldset";
import { FormBaseType } from "../../schema/blocks/form/utils";
import { FormWrapper } from "../components/FormWrapper";
import Input from "~/design/Input";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type SelectProps = BlockEditFormProps<FormSelectType>;

export const Select = ({
  action,
  block: propBlock,
  path,
  onSubmit,
  onCancel,
}: SelectProps) => {
  const { t } = useTranslation();
  const form = useForm<FormSelectType>({
    resolver: zodResolver(FormSelectSchema),
    defaultValues: propBlock,
  });

  return (
    <FormWrapper
      description={t(
        "direktivPage.blockEditor.blockForms.formPrimitives.select.description"
      )}
      form={form}
      block={propBlock}
      action={action}
      path={path}
      onSubmit={onSubmit}
      onCancel={onCancel}
    >
      <BaseForm form={form as unknown as UseFormReturn<FormBaseType>} />
      <Controller
        control={form.control}
        name="values"
        render={({ field }) => (
          <Fieldset
            label={t(
              "direktivPage.blockEditor.blockForms.formPrimitives.select.valuesLabel"
            )}
          >
            <ArrayForm
              defaultValue={field.value || []}
              onChange={field.onChange}
              emptyItem=""
              itemIsValid={(item) =>
                typeof item === "string" && item.length > 0
              }
              renderItem={({ value: itemValue, setValue, handleKeyDown }) => (
                <Input
                  placeholder={t(
                    "direktivPage.blockEditor.blockForms.formPrimitives.select.valuesPlaceholder"
                  )}
                  value={itemValue}
                  onKeyDown={handleKeyDown}
                  onChange={(e) => setValue(e.target.value)}
                />
              )}
            />
          </Fieldset>
        )}
      />
      <Fieldset
        label={t(
          "direktivPage.blockEditor.blockForms.formPrimitives.select.defaultValueLabel"
        )}
        htmlFor="defaultValue"
      >
        <Input
          {...form.register("defaultValue")}
          id="defaultValue"
          placeholder={t(
            "direktivPage.blockEditor.blockForms.formPrimitives.select.defaultValueLabelPlaceholder"
          )}
        />
      </Fieldset>
    </FormWrapper>
  );
};

import { Controller, ControllerRenderProps, useForm } from "react-hook-form";
import {
  DefaultValueTypeSchema,
  FormCheckbox,
  FormCheckboxType,
  allowedDefaultValueTypes,
} from "../../schema/blocks/form/checkbox";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import { BaseForm } from "./BaseForm";
import { BlockEditFormProps } from "..";
import { Checkbox as CheckboxDesignComponent } from "~/design/Checkbox";
import { Fieldset } from "~/components/Form/Fieldset";
import { FormWrapper } from "../components/FormWrapper";
import Input from "~/design/Input";
import { VariableInput } from "../components/VariableInput";
import { useTranslation } from "react-i18next";
import z from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type CheckboxProps = BlockEditFormProps<FormCheckboxType>;

export const Checkbox = ({
  action,
  block: propBlock,
  path,
  onSubmit,
  onCancel,
}: CheckboxProps) => {
  const { t } = useTranslation();
  const form = useForm<FormCheckboxType>({
    resolver: zodResolver(FormCheckbox),
    defaultValues: propBlock,
  });

  const onSelectChange = (
    field: ControllerRenderProps<FormCheckboxType, "defaultValue.type">,
    value: string
  ) => {
    const parsedValueType = DefaultValueTypeSchema.safeParse(value);
    if (parsedValueType.data) {
      field.onChange(parsedValueType.data);
      switch (parsedValueType.data) {
        case "boolean":
          form.setValue("defaultValue.value", false);
          break;
        case "variable":
          form.setValue("defaultValue.value", "");
          break;
      }
    }
  };

  return (
    <FormWrapper
      description={t(
        "direktivPage.blockEditor.blockForms.formPrimitives.checkbox.description"
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
          "direktivPage.blockEditor.blockForms.formPrimitives.defaultValue.label"
        )}
        htmlFor="defaultValue-type"
      >
        <div className="grid grid-cols-[110px,auto] items-center gap-2">
          <Controller
            control={form.control}
            name="defaultValue.type"
            render={({ field }) => (
              <Select
                value={field.value}
                onValueChange={(value) => {
                  onSelectChange(field, value);
                }}
              >
                <SelectTrigger variant="outline" id="defaultValue-type">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {allowedDefaultValueTypes.map((type) => (
                    <SelectItem value={type} key={type}>
                      {t(
                        `direktivPage.blockEditor.blockForms.formPrimitives.defaultValue.type.${type}`
                      )}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          />
          {form.watch("defaultValue.type") === "boolean" && (
            <Controller
              control={form.control}
              name="defaultValue.value"
              render={({ field }) => {
                const parsedValue = z.boolean().safeParse(field.value);
                const defaultValue = parsedValue.success
                  ? parsedValue.data
                  : false;

                return (
                  <CheckboxDesignComponent
                    checked={defaultValue}
                    onCheckedChange={(value) => {
                      if (value === "indeterminate") return;
                      field.onChange(value);
                    }}
                  />
                );
              }}
            />
          )}
          {form.watch("defaultValue.type") === "variable" && (
            <Controller
              control={form.control}
              name="defaultValue.value"
              render={({ field }) => {
                const parsedValue = z.string().safeParse(field.value);
                const defaultValue = parsedValue.success
                  ? parsedValue.data
                  : "";

                return (
                  <VariableInput
                    value={defaultValue}
                    onUpdate={(value) =>
                      form.setValue("defaultValue.value", value)
                    }
                    placeholder={t(
                      "direktivPage.blockEditor.blockForms.formPrimitives.defaultValue.placeholderVariable"
                    )}
                  />
                );
              }}
            />
          )}
        </div>
      </Fieldset>
    </FormWrapper>
  );
};

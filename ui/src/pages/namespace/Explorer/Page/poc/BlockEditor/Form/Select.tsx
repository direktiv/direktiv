import { Controller, ControllerRenderProps, useForm } from "react-hook-form";
import {
  FormSelect,
  FormSelectType,
  ValuesTypeSchema,
  allowedValuesTypes,
} from "../../schema/blocks/form/select";
import {
  SelectContent,
  Select as SelectDesignComponent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import { ArrayForm } from "~/components/Form/Array";
import { BaseForm } from "./BaseForm";
import { BlockEditFormProps } from "..";
import { Fieldset } from "~/components/Form/Fieldset";
import { FormWrapper } from "../components/FormWrapper";
import Input from "~/design/Input";
import { SmartInput } from "../components/SmartInput";
import { useTranslation } from "react-i18next";
import z from "zod";
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
    resolver: zodResolver(FormSelect),
    defaultValues: propBlock,
  });

  const onSelectChange = (
    field: ControllerRenderProps<FormSelectType, "values.type">,
    value: string
  ) => {
    const parsedValueType = ValuesTypeSchema.safeParse(value);
    if (parsedValueType.data) {
      field.onChange(parsedValueType.data);
      switch (parsedValueType.data) {
        case "array":
          form.setValue("values.value", []);
          break;
        case "variable":
          form.setValue("values.value", "");
          break;
      }
    }
  };

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
      <BaseForm form={form} />
      <Fieldset
        label={t(
          "direktivPage.blockEditor.blockForms.formPrimitives.select.valuesLabel"
        )}
        htmlFor="values-type"
      >
        <Controller
          control={form.control}
          name="values.type"
          render={({ field }) => (
            <SelectDesignComponent
              value={field.value}
              onValueChange={(value) => {
                onSelectChange(field, value);
              }}
            >
              <SelectTrigger variant="outline" id="values-type">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {allowedValuesTypes.map((type) => (
                  <SelectItem value={type} key={type}>
                    {t(
                      `direktivPage.blockEditor.blockForms.formPrimitives.select.type.${type}`
                    )}
                  </SelectItem>
                ))}
              </SelectContent>
            </SelectDesignComponent>
          )}
        />

        {form.watch("values.type") === "array" && (
          <Controller
            control={form.control}
            name="values.value"
            render={({ field }) => {
              const parsedValue = z.array(z.string()).safeParse(field.value);
              const defaultValue = parsedValue.success ? parsedValue.data : [];
              return (
                <ArrayForm
                  defaultValue={defaultValue}
                  onChange={field.onChange}
                  emptyItem=""
                  itemIsValid={(item) =>
                    typeof item === "string" && item.length > 0
                  }
                  renderItem={({ value, setValue, handleKeyDown }) => (
                    <Input
                      placeholder={t(
                        "direktivPage.blockEditor.blockForms.formPrimitives.select.valuesPlaceholder"
                      )}
                      value={value}
                      onKeyDown={handleKeyDown}
                      onChange={(e) => setValue(e.target.value)}
                    />
                  )}
                />
              );
            }}
          />
        )}
        {form.watch("values.type") === "variable" && (
          <Controller
            control={form.control}
            name="values.value"
            render={({ field }) => {
              const parsedValue = z.string().safeParse(field.value);
              const value = parsedValue.success ? parsedValue.data : "";
              return (
                <Input
                  {...field}
                  value={value}
                  placeholder={t(
                    "direktivPage.blockEditor.blockForms.formPrimitives.defaultValue.placeholderVariable"
                  )}
                />
              );
            }}
          />
        )}
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
            "direktivPage.blockEditor.blockForms.formPrimitives.defaultValue.placeholderSelect"
          )}
        />
      </Fieldset>
    </FormWrapper>
  );
};

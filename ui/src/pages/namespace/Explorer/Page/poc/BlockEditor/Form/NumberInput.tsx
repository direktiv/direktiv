import { Controller, ControllerRenderProps, useForm } from "react-hook-form";
import {
  DefaultValueTypeSchema,
  FormNumberInput,
  FormNumberInputType,
  allowedDefaultValueTypes,
} from "../../schema/blocks/form/numberInput";
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
import Input from "~/design/Input";
import { useTranslation } from "react-i18next";
import z from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type NumberInputProps = BlockEditFormProps<FormNumberInputType>;

export const NumberInput = ({
  action,
  block: propBlock,
  path,
  onSubmit,
  onCancel,
}: NumberInputProps) => {
  const { t } = useTranslation();
  const form = useForm<FormNumberInputType>({
    resolver: zodResolver(FormNumberInput),
    defaultValues: propBlock,
  });

  const onSelectChange = (
    field: ControllerRenderProps<FormNumberInputType, "defaultValue.type">,
    value: string
  ) => {
    const parsedValueType = DefaultValueTypeSchema.safeParse(value);
    if (parsedValueType.data) {
      field.onChange(parsedValueType.data);
      switch (parsedValueType.data) {
        case "number":
          form.setValue("defaultValue.value", 0);
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
        "direktivPage.blockEditor.blockForms.formPrimitives.numberInput.description"
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
          {form.watch("defaultValue.type") === "number" && (
            <Controller
              control={form.control}
              name="defaultValue.value"
              render={({ field }) => {
                const parsedValue = z.number().safeParse(field.value);
                const defaultValue = parsedValue.success ? parsedValue.data : 0;

                return (
                  <Input
                    {...field}
                    value={defaultValue}
                    onChange={(e) => {
                      const numValue = parseInt(e.target.value);
                      field.onChange(isNaN(numValue) ? 0 : numValue);
                    }}
                    type="number"
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
                  <Input
                    {...field}
                    value={defaultValue}
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

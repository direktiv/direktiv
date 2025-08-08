import { Controller, useForm } from "react-hook-form";
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
          "direktivPage.blockEditor.blockForms.formPrimitives.numberInput.defaultValueLabel"
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
                  const parsed = DefaultValueTypeSchema.safeParse(value);
                  if (parsed.success) {
                    field.onChange(parsed.data);
                    // reset value
                    if (parsed.data === "number") {
                      form.setValue("defaultValue.value", 0);
                    } else {
                      form.setValue("defaultValue.value", "");
                    }
                  }
                }}
              >
                <SelectTrigger variant="outline" id="defaultValue-type">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {allowedDefaultValueTypes.map((type) => (
                    <SelectItem value={type} key={type}>
                      {t(
                        `direktivPage.blockEditor.blockForms.formPrimitives.numberInput.defaultValueType.${type}`
                      )}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          />
          {form.watch("defaultValue.type") === "number" ? (
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
          ) : (
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
                      "direktivPage.blockEditor.blockForms.formPrimitives.numberInput.defaultValueVariablePlaceholder"
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

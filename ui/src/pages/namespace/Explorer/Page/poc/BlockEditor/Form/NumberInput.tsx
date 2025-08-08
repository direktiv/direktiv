import {
  FormNumberInput,
  FormNumberInputType,
} from "../../schema/blocks/form/numberInput";

import { BaseForm } from "./BaseForm";
import { BlockEditFormProps } from "..";
import { Fieldset } from "~/components/Form/Fieldset";
import { FormWrapper } from "../components/FormWrapper";
import InputDesignComponent from "~/design/Input";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
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
        htmlFor="defaultValue"
      >
        <InputDesignComponent
          {...form.register("defaultValue", { valueAsNumber: true })}
          id="defaultValue"
          type="number"
        />
      </Fieldset>
    </FormWrapper>
  );
};

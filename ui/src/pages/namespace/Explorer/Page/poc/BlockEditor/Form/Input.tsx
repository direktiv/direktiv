import {
  FormInput as FormInputSchema,
  FormInputType,
} from "../../schema/blocks/form/input";

import { BlockEditFormProps } from "..";
import { Checkbox } from "~/design/Checkbox";
import { Fieldset } from "~/components/Form/Fieldset";
import { FormWrapper } from "../components/FormWrapper";
import InputDesignComponent from "~/design/Input";
import { useForm } from "react-hook-form";
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
      <Fieldset
        label={t("direktivPage.blockEditor.blockForms.formPrimitives.idLabel")}
        htmlFor="id"
      >
        <InputDesignComponent
          {...form.register("id")}
          id="id"
          placeholder={t(
            "direktivPage.blockEditor.blockForms.formPrimitives.idPlaceholder"
          )}
        />
      </Fieldset>
      <Fieldset
        label={t("direktivPage.blockEditor.blockForms.formPrimitives.label")}
        htmlFor="label"
      >
        <InputDesignComponent
          {...form.register("label")}
          id="label"
          placeholder={t(
            "direktivPage.blockEditor.blockForms.formPrimitives.labelPlaceholder"
          )}
        />
      </Fieldset>
      <Fieldset
        label={t(
          "direktivPage.blockEditor.blockForms.formPrimitives.description"
        )}
        htmlFor="description"
      >
        <InputDesignComponent
          {...form.register("description")}
          id="description"
          placeholder={t(
            "direktivPage.blockEditor.blockForms.formPrimitives.descriptionPlaceholder"
          )}
        />
      </Fieldset>
      <Fieldset
        label={t("direktivPage.blockEditor.blockForms.formPrimitives.required")}
        htmlFor="required"
        horizontal
      >
        <Checkbox
          defaultChecked={form.getValues("required")}
          onCheckedChange={(value) => {
            if (typeof value === "boolean") {
              form.setValue("required", value);
            }
          }}
          id="required"
        />
      </Fieldset>
    </FormWrapper>
  );
};

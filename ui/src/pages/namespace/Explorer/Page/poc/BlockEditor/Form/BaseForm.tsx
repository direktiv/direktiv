import { Checkbox } from "~/design/Checkbox";
import { Fieldset } from "~/components/Form/Fieldset";
import { FormBaseType } from "../../schema/blocks/form/utils";
import InputDesignComponent from "~/design/Input";
import { UseFormReturn } from "react-hook-form";
import { useTranslation } from "react-i18next";

type BaseFormProps = {
  form: UseFormReturn<FormBaseType>;
};

export const BaseForm = ({ form }: BaseFormProps) => {
  const { t } = useTranslation();
  return (
    <>
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
    </>
  );
};

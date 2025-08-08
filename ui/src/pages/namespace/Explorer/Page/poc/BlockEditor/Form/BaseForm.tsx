import { Checkbox } from "~/design/Checkbox";
import { Fieldset } from "~/components/Form/Fieldset";
import { FormBaseType } from "../../schema/blocks/form/utils";
import InputDesignComponent from "~/design/Input";
import { UseFormReturn } from "react-hook-form";
import { useTranslation } from "react-i18next";

type BaseFormProps = {
  // Unfortunately, we cannot type form `type UseFormReturn<FormBaseType>` but have to use `UseFormReturn<any>`. Every form that we pass
  // to this component will implement the BaseForm but will still have additional properties. These additional properties will lead to a
  // type error. That's why we have to use `any` here. Still, this is a better solution that having boilerplate code in every form.
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  form: UseFormReturn<any>;
};
export const BaseForm = ({ form: anyForm }: BaseFormProps) => {
  const { t } = useTranslation();
  const form = anyForm as UseFormReturn<FormBaseType>;
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
        label={t("direktivPage.blockEditor.blockForms.formPrimitives.optional")}
        htmlFor="required"
        horizontal
      >
        <Checkbox
          defaultChecked={form.getValues("optional")}
          onCheckedChange={(value) => {
            if (typeof value === "boolean") {
              form.setValue("optional", value);
            }
          }}
          id="required"
        />
      </Fieldset>
    </>
  );
};

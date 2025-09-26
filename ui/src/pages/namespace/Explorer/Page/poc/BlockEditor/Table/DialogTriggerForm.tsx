import { Dialog, DialogType } from "../../schema/blocks/dialog";
import FormErrors, { errorsType } from "~/components/FormErrors";

import { Fieldset } from "~/components/Form/Fieldset";
import Input from "~/design/Input";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type DialogTriggerFormProps = {
  defaultValues?: DialogType;
  formId: string;
  onSubmit: (data: DialogType) => void;
};

export const DialogTriggerForm = ({
  defaultValues,
  formId,
  onSubmit,
}: DialogTriggerFormProps) => {
  const { t } = useTranslation();
  const form = useForm<DialogType>({
    resolver: zodResolver(Dialog),
    defaultValues: {
      type: "dialog",
      trigger: {
        type: "button",
        label: "",
      },
      blocks: [],
      ...defaultValues,
    },
  });

  const onFormSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.stopPropagation();
    form.handleSubmit(onSubmit)(e);
  };

  return (
    <form onSubmit={onFormSubmit} id={formId}>
      {form.formState.errors && (
        <FormErrors
          errors={form.formState.errors as errorsType}
          className="mb-5"
        />
      )}
      <Fieldset
        label={t(
          "direktivPage.blockEditor.blockForms.dialog.triggerLabelLabel"
        )}
        htmlFor="label"
      >
        <Input
          {...form.register("trigger.label")}
          id="label"
          placeholder={t(
            "direktivPage.blockEditor.blockForms.dialog.triggerLabelPlaceholder"
          )}
        />
      </Fieldset>
    </form>
  );
};

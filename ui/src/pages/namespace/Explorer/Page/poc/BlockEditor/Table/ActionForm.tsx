import { Dialog, DialogType } from "../../schema/blocks/dialog";
import FormErrors, { errorsType } from "~/components/FormErrors";

import { Fieldset } from "~/components/Form/Fieldset";
import { SmartInput } from "../components/SmartInput";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type ActionFormProps = {
  defaultValues?: DialogType;
  formId: string;
  onSubmit: (data: DialogType) => void;
};

// TODO:
// - may use the Fieldset from the dialog
// - rename this component
export const ActionForm = ({
  defaultValues,
  formId,
  onSubmit,
}: ActionFormProps) => {
  const { t } = useTranslation();
  const {
    watch,
    handleSubmit,
    setValue,
    formState: { errors },
  } = useForm<DialogType>({
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
    handleSubmit(onSubmit)(e);
  };

  return (
    <form onSubmit={onFormSubmit} id={formId}>
      {errors && <FormErrors errors={errors as errorsType} className="mb-5" />}
      <Fieldset
        label={t(
          "direktivPage.blockEditor.blockForms.dialog.triggerLabelLabel"
        )}
        htmlFor="label"
      >
        <SmartInput
          value={watch("trigger.label")}
          onUpdate={(value) => setValue("trigger.label", value)}
          id="label"
          placeholder={t(
            "direktivPage.blockEditor.blockForms.dialog.triggerLabelPlaceholder"
          )}
        />
      </Fieldset>
    </form>
  );
};

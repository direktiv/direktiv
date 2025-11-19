import { Dialog, DialogType } from "../../schema/blocks/dialog";
import FormErrors, { errorsType } from "~/components/FormErrors";

import { TriggerLabelFieldset } from "../components/TriggerLabelFieldset";
import { useForm } from "react-hook-form";
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
      <TriggerLabelFieldset form={form} />
    </form>
  );
};

import { Controller, useForm } from "react-hook-form";
import { Form as FormSchema, FormType } from "../../schema/blocks/form";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import { BlockEditFormProps } from "..";
import { ExtendedKeyValueInput } from "../components/FormElements/ExtendedKeyValueInput";
import { Fieldset } from "~/components/Form/Fieldset";
import { FormWrapper } from "../components/FormWrapper";
import { KeyValueInput } from "../components/FormElements/KeyValueInput";
import { SmartInput } from "../components/SmartInput";
import { mutationMethods } from "../../schema/procedures/mutation";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type FormEditFormProps = BlockEditFormProps<FormType>;

export const Form = ({
  action,
  block: propBlock,
  path,
  onSubmit,
  onCancel,
}: FormEditFormProps) => {
  const { t } = useTranslation();
  const form = useForm<FormType>({
    resolver: zodResolver(FormSchema),
    defaultValues: propBlock,
  });

  return (
    <FormWrapper
      description={t("direktivPage.blockEditor.blockForms.form.description")}
      form={form}
      block={propBlock}
      action={action}
      path={path}
      onSubmit={onSubmit}
      onCancel={onCancel}
    >
      <Fieldset
        label={t(
          "direktivPage.blockEditor.blockForms.form.mutation.triggerLabelLabel"
        )}
        htmlFor="trigger-label"
      >
        <SmartInput
          value={form.watch("trigger.label")}
          onUpdate={(value) => form.setValue("trigger.label", value)}
          id="trigger-label"
          placeholder={t(
            "direktivPage.blockEditor.blockForms.form.mutation.triggerLabelPlaceholder"
          )}
        />
      </Fieldset>
      <Fieldset
        label={t(
          "direktivPage.blockEditor.blockForms.form.mutation.methodLabel"
        )}
        htmlFor="mutation-method"
      >
        <Controller
          control={form.control}
          name="mutation.method"
          render={({ field }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger variant="outline" id="mutation-method">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {mutationMethods.map((method) => (
                  <SelectItem key={method} value={method}>
                    <span>{method}</span>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        />
      </Fieldset>
      <Fieldset
        label={t("direktivPage.blockEditor.blockForms.form.mutation.urlLabel")}
        htmlFor="mutation-url"
      >
        <SmartInput
          value={form.watch("mutation.url")}
          onUpdate={(value) => form.setValue("mutation.url", value)}
          id="mutation-url"
          placeholder={t(
            "direktivPage.blockEditor.blockForms.form.mutation.urlPlaceholder"
          )}
        />
      </Fieldset>
      <Controller
        control={form.control}
        name="mutation.queryParams"
        render={({ field }) => (
          <KeyValueInput
            field={field}
            label={t(
              "direktivPage.blockEditor.blockForms.form.mutation.queryParamsLabel"
            )}
          />
        )}
      />
      <Controller
        control={form.control}
        name="mutation.requestHeaders"
        render={({ field }) => (
          <KeyValueInput
            field={field}
            label={t(
              "direktivPage.blockEditor.blockForms.form.mutation.requestHeadersLabel"
            )}
          />
        )}
      />
      <Controller
        control={form.control}
        name="mutation.requestBody"
        render={({ field }) => (
          <ExtendedKeyValueInput
            field={field}
            label={t(
              "direktivPage.blockEditor.blockForms.form.mutation.requestBodyLabel"
            )}
          />
        )}
      />
    </FormWrapper>
  );
};

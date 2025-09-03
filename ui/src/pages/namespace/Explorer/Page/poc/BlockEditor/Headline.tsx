import { Controller, useForm } from "react-hook-form";
import {
  Headline as HeadlineSchema,
  HeadlineType,
  headlineLevels,
} from "../schema/blocks/headline";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import { BlockEditFormProps } from ".";
import { Fieldset } from "~/components/Form/Fieldset";
import { FormWrapper } from "./components/FormWrapper";
import { SmartInput } from "./components/SmartInput";
import { usePageEditorPanel } from "./EditorPanelProvider";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type HeadlineEditFormProps = BlockEditFormProps<HeadlineType>;

export const Headline = ({
  action,
  block: propBlock,
  path,
  onSubmit,
  onCancel,
}: HeadlineEditFormProps) => {
  const { t } = useTranslation();
  const form = useForm<HeadlineType>({
    resolver: zodResolver(HeadlineSchema),
    defaultValues: propBlock,
  });

  const { panel } = usePageEditorPanel();

  if (!panel) return null;

  return (
    <FormWrapper
      description={t(
        "direktivPage.blockEditor.blockForms.headline.description"
      )}
      form={form}
      block={propBlock}
      action={action}
      path={path}
      onSubmit={onSubmit}
      onCancel={onCancel}
    >
      <Fieldset
        label={t("direktivPage.blockEditor.blockForms.headline.labelLabel")}
        htmlFor="label"
      >
        <SmartInput
          value={form.watch("label")}
          onChange={(content) => form.setValue("label", content)}
          id="label"
          variables={panel.variables}
          placeholder={t(
            "direktivPage.blockEditor.blockForms.headline.labelPlaceholder"
          )}
        />
      </Fieldset>
      <Fieldset
        label={t("direktivPage.blockEditor.blockForms.headline.levelLabel")}
        htmlFor="level"
      >
        <Controller
          control={form.control}
          name="level"
          render={({ field }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger variant="outline" id="level">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {headlineLevels.map((item) => (
                  <SelectItem key={item} value={item}>
                    <span>{item}</span>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        />
      </Fieldset>
    </FormWrapper>
  );
};

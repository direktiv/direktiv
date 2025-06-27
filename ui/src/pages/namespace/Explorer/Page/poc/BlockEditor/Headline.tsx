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
import Input from "~/design/Input";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type HeadlineEditFormProps = BlockEditFormProps<HeadlineType>;

export const Headline = ({
  block: propBlock,
  onSubmit,
}: HeadlineEditFormProps) => {
  const { t } = useTranslation();
  const form = useForm<HeadlineType>({
    resolver: zodResolver(HeadlineSchema),
    defaultValues: propBlock,
  });

  return (
    <FormWrapper form={form} onSubmit={onSubmit}>
      <div className="text-gray-10 dark:text-gray-10">
        {t("direktivPage.blockEditor.blockForms.headline.description")}
      </div>
      <Fieldset
        label={t("direktivPage.blockEditor.blockForms.headline.labelLabel")}
        htmlFor="label"
      >
        <Input
          {...form.register("label")}
          id="label"
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

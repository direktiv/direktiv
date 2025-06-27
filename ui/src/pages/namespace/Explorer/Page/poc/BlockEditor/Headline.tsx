import { Controller, useForm } from "react-hook-form";
import FormErrors, { errorsType } from "~/components/FormErrors";
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
import { DialogFooter } from "./components/Footer";
import { DialogHeader } from "./components/Header";
import { Fieldset } from "~/components/Form/Fieldset";
import Input from "~/design/Input";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type HeadlineEditFormProps = BlockEditFormProps<HeadlineType>;

const formId = "block-editor-headline";

export const Headline = ({
  action,
  block: propBlock,
  path,
  onSubmit,
}: HeadlineEditFormProps) => {
  const { t } = useTranslation();
  const {
    handleSubmit,
    register,
    control,
    formState: { errors },
  } = useForm<HeadlineType>({
    resolver: zodResolver(HeadlineSchema),
    defaultValues: propBlock,
  });

  return (
    <form
      onSubmit={handleSubmit(onSubmit)}
      id={formId}
      className="flex flex-col gap-3"
    >
      <DialogHeader action={action} path={path} type={propBlock.type} />
      {errors && <FormErrors errors={errors as errorsType} />}
      <div className="text-gray-10 dark:text-gray-10">
        {t("direktivPage.blockEditor.blockForms.headline.description")}
      </div>
      <Fieldset
        label={t("direktivPage.blockEditor.blockForms.headline.labelLabel")}
        htmlFor="label"
      >
        <Input
          {...register("label")}
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
          control={control}
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
      <DialogFooter formId={formId} />
    </form>
  );
};

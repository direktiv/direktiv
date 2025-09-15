import { Image as ImageSchema, ImageType } from "../schema/blocks/image";

import { BlockEditFormProps } from ".";
import { Fieldset } from "~/components/Form/Fieldset";
import { FormWrapper } from "./components/FormWrapper";
import Input from "~/design/Input";
import { SmartInput } from "./components/SmartInput";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type ImageEditFormProps = BlockEditFormProps<ImageType>;

export const Image = ({
  action,
  block: propBlock,
  path,
  onSubmit,
  onCancel,
}: ImageEditFormProps) => {
  const { t } = useTranslation();
  const form = useForm<ImageType>({
    resolver: zodResolver(ImageSchema),
    defaultValues: propBlock,
  });

  return (
    <FormWrapper
      description={t("direktivPage.blockEditor.blockForms.image.description")}
      form={form}
      block={propBlock}
      action={action}
      path={path}
      onSubmit={onSubmit}
      onCancel={onCancel}
    >
      <Fieldset
        label={t("direktivPage.blockEditor.blockForms.image.srcLabel")}
        htmlFor="src"
      >
        <SmartInput
          value={form.watch("src")}
          onUpdate={(value) => form.setValue("src", value)}
          id="src"
          placeholder={t(
            "direktivPage.blockEditor.blockForms.image.srcPlaceholder"
          )}
        />
      </Fieldset>
      <Fieldset
        label={t("direktivPage.blockEditor.blockForms.image.widthLabel")}
        htmlFor="width"
      >
        <Input
          {...form.register("width", { valueAsNumber: true })}
          id="width"
          type="number"
          placeholder={t(
            "direktivPage.blockEditor.blockForms.image.widthPlaceholder"
          )}
        />
      </Fieldset>
      <Fieldset
        label={t("direktivPage.blockEditor.blockForms.image.heightLabel")}
        htmlFor="height"
      >
        <Input
          {...form.register("height", { valueAsNumber: true })}
          id="height"
          type="number"
          placeholder={t(
            "direktivPage.blockEditor.blockForms.image.heightPlaceholder"
          )}
        />
      </Fieldset>
    </FormWrapper>
  );
};

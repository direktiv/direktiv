import { Loop as LoopSchema, LoopType } from "../schema/blocks/loop";

import { BlockEditFormProps } from ".";
import { Fieldset } from "~/components/Form/Fieldset";
import { FormWrapper } from "./components/FormWrapper";
import Input from "~/design/Input";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type LoopFormProps = BlockEditFormProps<LoopType>;
//
export const Loop = ({
  action,
  block: propBlock,
  path,
  onSubmit,
}: LoopFormProps) => {
  const { t } = useTranslation();
  const form = useForm<LoopType>({
    resolver: zodResolver(LoopSchema),
    defaultValues: propBlock,
  });

  return (
    <FormWrapper
      description={t("direktivPage.blockEditor.blockForms.loop.description")}
      form={form}
      onSubmit={onSubmit}
      action={action}
      path={path}
      blockType={propBlock.type}
    >
      <Fieldset
        label={t("direktivPage.blockEditor.blockForms.loop.idLabel")}
        htmlFor="id"
      >
        <Input
          {...form.register("id")}
          id="id"
          placeholder={t(
            "direktivPage.blockEditor.blockForms.loop.idPlaceholder"
          )}
        />
      </Fieldset>
      <Fieldset
        label={t("direktivPage.blockEditor.blockForms.loop.dataLabel")}
        htmlFor="data"
      >
        <Input
          {...form.register("data")}
          id="data"
          placeholder={t(
            "direktivPage.blockEditor.blockForms.loop.dataPlaceholder"
          )}
        />
      </Fieldset>
    </FormWrapper>
  );
};

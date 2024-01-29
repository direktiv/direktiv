import { Select, SelectTrigger, SelectValue } from "~/design/Select";

import { FC } from "react";
import { Fieldset } from "~/components/Form/Fieldset";
import Input from "~/design/Input";
import { useTranslation } from "react-i18next";

type PatchFormProps = {};

export const PatchForm: FC<PatchFormProps> = () => {
  const { t } = useTranslation();

  return (
    <div>
      <Fieldset
        label={t("pages.explorer.service.editor.form.patches.modal.op.label")}
        htmlFor="op"
      >
        <Select>
          <SelectTrigger id="op" variant="outline">
            <SelectValue
              placeholder={t(
                "pages.explorer.service.editor.form.patches.modal.op.placeholder"
              )}
            />
          </SelectTrigger>
        </Select>
      </Fieldset>
      <Fieldset
        label={t("pages.explorer.service.editor.form.patches.modal.path")}
        htmlFor="path"
      >
        <Input type="text" id="path" />
      </Fieldset>
      <Fieldset
        label={t("pages.explorer.service.editor.form.patches.modal.value")}
        htmlFor="value"
      >
        <Input type="text" id="value" />
      </Fieldset>
    </div>
  );
};

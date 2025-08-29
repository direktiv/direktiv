import {
  ExtendedKeyValue,
  ExtendedKeyValueType,
} from "../../../../schema/primitives/extendedKeyValue";

import { ArrayForm } from "~/components/Form/Array";
import { Card } from "~/design/Card";
import { Fieldset } from "~/components/Form/Fieldset";
import Input from "~/design/Input";
import { PropsWithChildren } from "react";
import { ValueInput } from "./ValueInput";
import { useTranslation } from "react-i18next";

type ExtendedKeyValueInputProps = {
  label: string;
  field: {
    value: ExtendedKeyValueType[] | undefined;
    onChange: (value: ExtendedKeyValueType[]) => void;
  };
};

export const ExtendedKeyValueInput = ({
  field,
  label,
}: ExtendedKeyValueInputProps) => {
  const { t } = useTranslation();
  return (
    <Fieldset label={label}>
      <div className="flex flex-col gap-y-5">
        <ArrayForm
          defaultValue={field.value || []}
          onChange={field.onChange}
          emptyItem={{
            key: "",
            value: {
              type: "string",
              value: "",
            },
          }}
          itemIsValid={(item) => ExtendedKeyValue.safeParse(item).success}
          wrapItem={(children) => (
            <Card className="flex flex-col gap-3 p-3" noShadow>
              {children}
            </Card>
          )}
          renderItem={({ value: itemValue, setValue, handleKeyDown }) => (
            <>
              <Container>
                <Label>
                  {t("direktivPage.blockEditor.blockForms.keyValue.key")}
                </Label>
                <Input
                  className="w-full"
                  placeholder={t(
                    "direktivPage.blockEditor.blockForms.keyValue.key"
                  )}
                  value={itemValue.key}
                  onKeyDown={handleKeyDown}
                  onChange={(e) => {
                    setValue({
                      ...itemValue,
                      key: e.target.value,
                    });
                  }}
                />
              </Container>
              <Container>
                <Label>
                  {t("direktivPage.blockEditor.blockForms.keyValue.value")}
                </Label>
                <div className="flex w-full items-center gap-2">
                  <ValueInput
                    value={itemValue.value}
                    onChange={(newValue) => {
                      setValue({
                        ...itemValue,
                        value: newValue,
                      });
                    }}
                    onKeyDown={handleKeyDown}
                  />
                </div>
              </Container>
            </>
          )}
        />
      </div>
    </Fieldset>
  );
};

const Label = ({ children }: PropsWithChildren) => (
  <label className="w-[55px] grow text-sm">{children}</label>
);

const Container = ({ children }: PropsWithChildren) => (
  <div className="flex items-center gap-2">{children}</div>
);
